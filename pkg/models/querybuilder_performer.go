package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/stashapp/stashdb/pkg/database"
	"github.com/stashapp/stashdb/pkg/utils"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type PerformerQueryBuilder struct {
	dbi database.DBI
}

func NewPerformerQueryBuilder(tx *sqlx.Tx) PerformerQueryBuilder {
	return PerformerQueryBuilder{
		dbi: database.DBIWithTxn(tx),
	}
}

func (qb *PerformerQueryBuilder) toModel(ro interface{}) *Performer {
	if ro != nil {
		return ro.(*Performer)
	}

	return nil
}

func (qb *PerformerQueryBuilder) Create(newPerformer Performer) (*Performer, error) {
	ret, err := qb.dbi.Insert(newPerformer)
	return qb.toModel(ret), err
}

func (qb *PerformerQueryBuilder) Update(updatedPerformer Performer) (*Performer, error) {
	ret, err := qb.dbi.Update(updatedPerformer, true)
	return qb.toModel(ret), err
}

func (qb *PerformerQueryBuilder) UpdatePartial(updatedPerformer Performer) (*Performer, error) {
	ret, err := qb.dbi.Update(updatedPerformer, false)
	return qb.toModel(ret), err
}

func (qb *PerformerQueryBuilder) Destroy(id uuid.UUID) error {
	return qb.dbi.Delete(id, performerDBTable)
}

func (qb *PerformerQueryBuilder) CreateAliases(newJoins PerformerAliases) error {
	return qb.dbi.InsertJoins(performerAliasTable, &newJoins)
}

func (qb *PerformerQueryBuilder) UpdateAliases(performerID uuid.UUID, updatedJoins PerformerAliases) error {
	return qb.dbi.ReplaceJoins(performerAliasTable, performerID, &updatedJoins)
}

func (qb *PerformerQueryBuilder) CreateUrls(newJoins PerformerUrls) error {
	return qb.dbi.InsertJoins(performerUrlTable, &newJoins)
}

func (qb *PerformerQueryBuilder) UpdateUrls(performerID uuid.UUID, updatedJoins PerformerUrls) error {
	return qb.dbi.ReplaceJoins(performerUrlTable, performerID, &updatedJoins)
}

func (qb *PerformerQueryBuilder) CreateTattoos(newJoins PerformerBodyMods) error {
	return qb.dbi.InsertJoins(performerTattooTable, &newJoins)
}

func (qb *PerformerQueryBuilder) UpdateTattoos(performerID uuid.UUID, updatedJoins PerformerBodyMods) error {
	return qb.dbi.ReplaceJoins(performerTattooTable, performerID, &updatedJoins)
}

func (qb *PerformerQueryBuilder) CreatePiercings(newJoins PerformerBodyMods) error {
	return qb.dbi.InsertJoins(performerPiercingTable, &newJoins)
}

func (qb *PerformerQueryBuilder) UpdatePiercings(performerID uuid.UUID, updatedJoins PerformerBodyMods) error {
	return qb.dbi.ReplaceJoins(performerPiercingTable, performerID, &updatedJoins)
}

func (qb *PerformerQueryBuilder) Find(id uuid.UUID) (*Performer, error) {
	ret, err := qb.dbi.Find(id, performerDBTable)
	return qb.toModel(ret), err
}

func (qb *PerformerQueryBuilder) FindByIds(ids []uuid.UUID) ([]*Performer, []error) {
	query := "SELECT performers.* FROM performers WHERE id IN (?)"
	query, args, _ := sqlx.In(query, ids)
	performers, err := qb.queryPerformers(query, args)
	if err != nil {
		return nil, utils.DuplicateError(err, len(ids))
	}

	m := make(map[uuid.UUID]*Performer)
	for _, performer := range performers {
		m[performer.ID] = performer
	}

	result := make([]*Performer, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func (qb *PerformerQueryBuilder) FindBySceneID(sceneID uuid.UUID) (Performers, error) {
	query := `
		SELECT performers.* FROM performers
		LEFT JOIN performers_scenes as scenes_join on scenes_join.performer_id = performers.id
		WHERE scenes_join.scene_id = ?
		GROUP BY performers.id
	`
	args := []interface{}{sceneID}
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) FindByNames(names []string) (Performers, error) {
	query := "SELECT * FROM performers WHERE name IN " + getInBinding(len(names))
	var args []interface{}
	for _, name := range names {
		args = append(args, name)
	}
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) FindByAliases(names []string) (Performers, error) {
	query := `SELECT performers.* FROM performers
		left join performer_aliases on performers.id = performer_aliases.performer_id
		WHERE performer_aliases.alias IN ` + getInBinding(len(names))

	var args []interface{}
	for _, name := range names {
		args = append(args, name)
	}
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) FindByName(name string) (Performers, error) {
	query := "SELECT * FROM performers WHERE upper(name) = upper(?)"
	var args []interface{}
	args = append(args, name)
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) FindByAlias(name string) (Performers, error) {
	query := `SELECT performers.* FROM performers
		left join performer_aliases on performers.id = performer_aliases.performer_id
		WHERE upper(performer_aliases.alias) = UPPER(?)`

	var args []interface{}
	args = append(args, name)
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) Count() (int, error) {
	return runCountQuery(buildCountQuery("SELECT performers.id FROM performers"), nil)
}

func (qb *PerformerQueryBuilder) Query(performerFilter *PerformerFilterType, findFilter *QuerySpec) ([]*Performer, int) {
	if performerFilter == nil {
		performerFilter = &PerformerFilterType{}
	}
	if findFilter == nil {
		findFilter = &QuerySpec{}
	}

	query := database.NewQueryBuilder(performerDBTable)
	query.Eq("deleted", false)

	if q := performerFilter.Name; q != nil && *q != "" {
		searchColumns := []string{"performers.name"}
		clause, thisArgs := getSearchBinding(searchColumns, *q, false, false)
		query.AddWhere(clause)
		query.AddArg(thisArgs...)
	}

	if birthYear := performerFilter.BirthYear; birthYear != nil {
		clauses, thisArgs := getBirthYearFilterClause(birthYear.Modifier, birthYear.Value)
		query.AddWhere(clauses...)
		query.AddArg(thisArgs...)
	}

	if age := performerFilter.Age; age != nil {
		clauses, thisArgs := getAgeFilterClause(age.Modifier, age.Value)
		query.AddWhere(clauses...)
		query.AddArg(thisArgs...)
	}

	//handleStringCriterion("ethnicity", performerFilter.Ethnicity, &query)
	handleStringCriterion("country", performerFilter.Country, query)
	//handleStringCriterion("eye_color", performerFilter.EyeColor, &query)
	//handleStringCriterion("height", performerFilter.Height, &query)
	//handleStringCriterion("measurements", performerFilter.Measurements, &query)
	//handleStringCriterion("fake_tits", performerFilter.FakeTits, &query)
	//handleStringCriterion("career_length", performerFilter.CareerLength, &query)
	//handleStringCriterion("tattoos", performerFilter.Tattoos, &query)
	//handleStringCriterion("piercings", performerFilter.Piercings, &query)
	//handleStringCriterion("aliases", performerFilter.Aliases, &query)

	query.SortAndPagination = qb.getPerformerSort(findFilter) + getPagination(findFilter)
	var performers Performers
	countResult, err := qb.dbi.Query(*query, &performers)

	if err != nil {
		// TODO
		panic(err)
	}

	return performers, countResult
}

func getBirthYearFilterClause(criterionModifier CriterionModifier, value int) ([]string, []interface{}) {
	var clauses []string
	var args []interface{}

	yearStr := strconv.Itoa(value)
	startOfYear := yearStr + "-01-01"
	endOfYear := yearStr + "-12-31"

	if modifier := criterionModifier.String(); criterionModifier.IsValid() {
		switch modifier {
		case "EQUALS":
			// between yyyy-01-01 and yyyy-12-31
			clauses = append(clauses, "performers.birthdate >= ?")
			clauses = append(clauses, "performers.birthdate <= ?")
			args = append(args, startOfYear)
			args = append(args, endOfYear)
		case "NOT_EQUALS":
			// outside of yyyy-01-01 to yyyy-12-31
			clauses = append(clauses, "performers.birthdate < ? OR performers.birthdate > ?")
			args = append(args, startOfYear)
			args = append(args, endOfYear)
		case "GREATER_THAN":
			// > yyyy-12-31
			clauses = append(clauses, "performers.birthdate > ?")
			args = append(args, endOfYear)
		case "LESS_THAN":
			// < yyyy-01-01
			clauses = append(clauses, "performers.birthdate < ?")
			args = append(args, startOfYear)
		}
	}

	return clauses, args
}

func getAgeFilterClause(criterionModifier CriterionModifier, value int) ([]string, []interface{}) {
	var clauses []string
	var args []interface{}

	// get the date at which performer would turn the age specified
	dt := time.Now()
	birthDate := dt.AddDate(-value-1, 0, 0)
	yearAfter := birthDate.AddDate(1, 0, 0)

	if modifier := criterionModifier.String(); criterionModifier.IsValid() {
		switch modifier {
		case "EQUALS":
			// between birthDate and yearAfter
			clauses = append(clauses, "performers.birthdate >= ?")
			clauses = append(clauses, "performers.birthdate < ?")
			args = append(args, birthDate)
			args = append(args, yearAfter)
		case "NOT_EQUALS":
			// outside of birthDate and yearAfter
			clauses = append(clauses, "performers.birthdate < ? OR performers.birthdate >= ?")
			args = append(args, birthDate)
			args = append(args, yearAfter)
		case "GREATER_THAN":
			// < birthDate
			clauses = append(clauses, "performers.birthdate < ?")
			args = append(args, birthDate)
		case "LESS_THAN":
			// > yearAfter
			clauses = append(clauses, "performers.birthdate >= ?")
			args = append(args, yearAfter)
		}
	}

	return clauses, args
}

func (qb *PerformerQueryBuilder) getPerformerSort(findFilter *QuerySpec) string {
	var sort string
	var direction string
	if findFilter == nil {
		sort = "name"
		direction = "ASC"
	} else {
		sort = findFilter.GetSort("name")
		direction = findFilter.GetDirection()
	}
	return getSort(sort, direction, "performers")
}

func (qb *PerformerQueryBuilder) queryPerformers(query string, args []interface{}) (Performers, error) {
	output := Performers{}
	err := qb.dbi.RawQuery(performerDBTable, query, args, &output)
	return output, err
}

func (qb *PerformerQueryBuilder) GetAliases(id uuid.UUID) (PerformerAliases, error) {
	joins := PerformerAliases{}
	err := qb.dbi.FindJoins(performerAliasTable, id, &joins)

	return joins, err
}

func (qb *PerformerQueryBuilder) GetAllAliases(ids []uuid.UUID) ([][]string, []error) {
	joins := PerformerAliases{}
	err := qb.dbi.FindAllJoins(performerAliasTable, ids, &joins)
	if err != nil {
		return nil, utils.DuplicateError(err, len(ids))
	}

	m := make(map[uuid.UUID][]string)
	for _, join := range joins {
		m[join.PerformerID] = append(m[join.PerformerID], join.Alias)
	}

	result := make([][]string, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func (qb *PerformerQueryBuilder) GetUrls(id uuid.UUID) ([]*URL, error) {
	joins := PerformerUrls{}
	err := qb.dbi.FindJoins(performerUrlTable, id, &joins)

	urls := make([]*URL, len(joins))
	for i, u := range joins {
		url := URL{
			URL:  u.URL,
			Type: u.Type,
		}
		urls[i] = &url
	}

	return urls, err
}

func (qb *PerformerQueryBuilder) GetAllUrls(ids []uuid.UUID) ([][]*URL, []error) {
	joins := PerformerUrls{}
	err := qb.dbi.FindAllJoins(performerUrlTable, ids, &joins)
	if err != nil {
		return nil, utils.DuplicateError(err, len(ids))
	}

	m := make(map[uuid.UUID][]*URL)
	for _, join := range joins {
		url := URL{
			URL:  join.URL,
			Type: join.Type,
		}
		m[join.PerformerID] = append(m[join.PerformerID], &url)
	}

	result := make([][]*URL, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func (qb *PerformerQueryBuilder) GetTattoos(id uuid.UUID) (PerformerBodyMods, error) {
	joins := PerformerBodyMods{}
	err := qb.dbi.FindJoins(performerTattooTable, id, &joins)

	return joins, err
}

func (qb *PerformerQueryBuilder) GetAllTattoos(ids []uuid.UUID) ([][]*BodyModification, []error) {
	joins := PerformerBodyMods{}
	err := qb.dbi.FindAllJoins(performerTattooTable, ids, &joins)
	if err != nil {
		return nil, utils.DuplicateError(err, len(ids))
	}

	m := make(map[uuid.UUID][]*BodyModification)
	for _, join := range joins {
		desc := &join.Description.String
		if !join.Description.Valid {
			desc = nil
		}
		mod := BodyModification{
			Location:    join.Location,
			Description: desc,
		}
		m[join.PerformerID] = append(m[join.PerformerID], &mod)
	}

	result := make([][]*BodyModification, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func (qb *PerformerQueryBuilder) GetPiercings(id uuid.UUID) (PerformerBodyMods, error) {
	joins := PerformerBodyMods{}
	err := qb.dbi.FindJoins(performerPiercingTable, id, &joins)

	return joins, err
}

func (qb *PerformerQueryBuilder) GetAllPiercings(ids []uuid.UUID) ([][]*BodyModification, []error) {
	joins := PerformerBodyMods{}
	err := qb.dbi.FindAllJoins(performerPiercingTable, ids, &joins)
	if err != nil {
		return nil, utils.DuplicateError(err, len(ids))
	}

	m := make(map[uuid.UUID][]*BodyModification)
	for _, join := range joins {
		desc := &join.Description.String
		if !join.Description.Valid {
			desc = nil
		}
		mod := BodyModification{
			Location:    join.Location,
			Description: desc,
		}
		m[join.PerformerID] = append(m[join.PerformerID], &mod)
	}

	result := make([][]*BodyModification, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func (qb *PerformerQueryBuilder) SearchPerformers(term string) (Performers, error) {
	query := `
        SELECT * FROM performers
        WHERE name % $1
        AND similarity(name, $1) > 0.5
        ORDER BY similarity(name, $1) DESC
        LIMIT 5`
	args := []interface{}{term}
	return qb.queryPerformers(query, args)
}

func (qb *PerformerQueryBuilder) DeleteScenePerformers(id uuid.UUID) error {
	// Delete scene_performers joins
	return qb.dbi.DeleteJoins(performerSceneTable, id)
}

func (qb *PerformerQueryBuilder) SoftDelete(performer Performer) (*Performer, error) {
	// Delete joins
	if err := qb.dbi.DeleteJoins(performerAliasTable, performer.ID); err != nil {
		return nil, err
	}
	if err := qb.dbi.DeleteJoins(performerPiercingTable, performer.ID); err != nil {
		return nil, err
	}
	if err := qb.dbi.DeleteJoins(performerTattooTable, performer.ID); err != nil {
		return nil, err
	}
	if err := qb.dbi.DeleteJoins(performerUrlTable, performer.ID); err != nil {
		return nil, err
	}

	ret, err := qb.dbi.SoftDelete(performer)
	return qb.toModel(ret), err
}

func (qb *PerformerQueryBuilder) CreateRedirect(newJoin PerformerRedirect) error {
	return qb.dbi.InsertJoin(performerRedirectTable, newJoin, false)
}

func (qb *PerformerQueryBuilder) UpdateRedirects(oldTargetID uuid.UUID, newTargetID uuid.UUID) error {
	query := "UPDATE " + performerRedirectTable.Table.Name() + " SET target_id = ? WHERE target_id = ?"
	args := []interface{}{newTargetID, oldTargetID}
	return qb.dbi.RawQuery(performerRedirectTable.Table, query, args, nil)
}

func (qb *PerformerQueryBuilder) UpdateScenePerformers(oldTargetID uuid.UUID, newTargetID uuid.UUID) error {
	// Insert new performers for any scenes that have the old performers
	query := `INSERT INTO scene_performers (scene_id, performer_id)
            SELECT scene_id, ? 
            FROM scene_performers WHERE performer_id = ?
            ON CONFLICT DO NOTHING`
	args := []interface{}{newTargetID, oldTargetID}
	err := qb.dbi.RawQuery(scenePerformerTable.Table, query, args, nil)
	if err != nil {
		return err
	}

	// Delete any joins with the old performer
	query = `DELETE FROM scene_performers WHERE performer_id = ?`
	args = []interface{}{oldTargetID}
	return qb.dbi.RawQuery(scenePerformerTable.Table, query, args, nil)
}

func (qb *PerformerQueryBuilder) MergeInto(sourceID uuid.UUID, targetID uuid.UUID) error {
	performer, err := qb.Find(sourceID)
	if err != nil {
		return err
	}
	if performer == nil {
		return errors.New("Merge source performer not found: " + sourceID.String())
	}
	if performer.Deleted {
		return errors.New("Merge source performer is deleted: " + sourceID.String())
	}
	_, err = qb.SoftDelete(*performer)
	if err != nil {
		return err
	}
	if err := qb.UpdateRedirects(sourceID, targetID); err != nil {
		return err
	}
	if err := qb.UpdateScenePerformers(sourceID, targetID); err != nil {
		return err
	}
	redirect := PerformerRedirect{SourceID: sourceID, TargetID: targetID}
	return qb.CreateRedirect(redirect)
}

func (qb *PerformerQueryBuilder) ApplyEdit(edit Edit, operation OperationEnum, performer *Performer) (*Performer, error) {
	data, err := edit.GetPerformerData()
	if err != nil {
		return nil, err
	}

	switch operation {
	case OperationEnumCreate:
		now := time.Now()
		UUID, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		newPerformer := Performer{
			ID:        UUID,
			CreatedAt: SQLiteTimestamp{Timestamp: now},
		}
		if data.New.Name == nil {
			return nil, errors.New("Missing performer name")
		}
		newPerformer.CopyFromPerformerEdit(*data.New)

		performer, err = qb.Create(newPerformer)
		if err != nil {
			return nil, err
		}

		if len(data.New.AddedAliases) > 0 {
			aliases := CreatePerformerAliases(UUID, data.New.AddedAliases)
			if err := qb.CreateAliases(aliases); err != nil {
				return nil, err
			}
		}

		if len(data.New.AddedTattoos) > 0 {
			tattoos := CreatePerformerBodyMods(UUID, data.New.AddedTattoos)
			if err := qb.CreateTattoos(tattoos); err != nil {
				return nil, err
			}
		}

		if len(data.New.AddedPiercings) > 0 {
			piercings := CreatePerformerBodyMods(UUID, data.New.AddedPiercings)
			if err := qb.CreatePiercings(piercings); err != nil {
				return nil, err
			}
		}

		if len(data.New.AddedUrls) > 0 {
			urls := CreatePerformerUrls(UUID, data.New.AddedUrls)
			if err := qb.CreateUrls(urls); err != nil {
				return nil, err
			}
		}

		// TODO
		//if len(data.New.AddedImages) > 0 {
		//images := CreatePerformerImages(UUID, data.New.AddedImages)
		//if err := qb.CreateImages(images); err != nil {
		//return nil, err
		//}
		//}

		return performer, nil
	case OperationEnumDestroy:
		updatedPerformer, err := qb.SoftDelete(*performer)
		if err != nil {
			return nil, err
		}
		err = qb.DeleteScenePerformers(performer.ID)
		return updatedPerformer, err
	case OperationEnumModify:
		if err := performer.ValidateModifyEdit(*data); err != nil {
			return nil, err
		}

		performer.CopyFromPerformerEdit(*data.New)
		updatedPerformer, err := qb.UpdatePartial(*performer)

		currentAliases, err := qb.GetAliases(updatedPerformer.ID)
		if err != nil {
			return nil, err
		}
		newAliases := CreatePerformerAliases(updatedPerformer.ID, data.New.AddedAliases)
		oldAliases := CreatePerformerAliases(updatedPerformer.ID, data.New.RemovedAliases)
		if err := ProcessSlice(&currentAliases, &newAliases, &oldAliases); err != nil {
			return nil, err
		}
		if err := qb.UpdateAliases(updatedPerformer.ID, currentAliases); err != nil {
			return nil, err
		}

		currentTattoos, err := qb.GetTattoos(updatedPerformer.ID)
		if err != nil {
			return nil, err
		}
		newTattoos := CreatePerformerBodyMods(updatedPerformer.ID, data.New.AddedTattoos)
		oldTattoos := CreatePerformerBodyMods(updatedPerformer.ID, data.New.RemovedTattoos)

		if err := ProcessSlice(&currentTattoos, &newTattoos, &oldTattoos); err != nil {
			return nil, err
		}
		if err := qb.UpdateTattoos(updatedPerformer.ID, currentTattoos); err != nil {
			return nil, err
		}

		currentPiercings, err := qb.GetPiercings(updatedPerformer.ID)
		if err != nil {
			return nil, err
		}
		newPiercings := CreatePerformerBodyMods(updatedPerformer.ID, data.New.AddedPiercings)
		oldPiercings := CreatePerformerBodyMods(updatedPerformer.ID, data.New.RemovedPiercings)

		if err := ProcessSlice(&currentPiercings, &newPiercings, &oldPiercings); err != nil {
			return nil, err
		}
		if err := qb.UpdatePiercings(updatedPerformer.ID, currentPiercings); err != nil {
			return nil, err
		}

		urls, err := qb.GetUrls(updatedPerformer.ID)
		currentUrls := CreatePerformerUrls(updatedPerformer.ID, urls)
		if err != nil {
			return nil, err
		}
		newUrls := CreatePerformerUrls(updatedPerformer.ID, data.New.AddedUrls)
		oldUrls := CreatePerformerUrls(updatedPerformer.ID, data.New.RemovedUrls)

		if err := ProcessSlice(&currentUrls, &newUrls, &oldUrls); err != nil {
			return nil, err
		}

		if err := qb.UpdateUrls(updatedPerformer.ID, currentUrls); err != nil {
			return nil, err
		}

		return updatedPerformer, err
	case OperationEnumMerge:
		if err := performer.ValidateModifyEdit(*data); err != nil {
			return nil, err
		}

		performer.CopyFromPerformerEdit(*data.New)
		updatedPerformer, err := qb.Update(*performer)

		for _, v := range data.MergeSources {
			sourceUUID, _ := uuid.FromString(v)
			if err := qb.MergeInto(sourceUUID, performer.ID); err != nil {
				return nil, err
			}
		}

		currentAliases, err := qb.GetAliases(updatedPerformer.ID)
		if err != nil {
			return nil, err
		}
		newAliases := CreatePerformerAliases(updatedPerformer.ID, data.New.AddedAliases)
		if err := currentAliases.AddAliases(newAliases); err != nil {
			return nil, err
		}
		if err := currentAliases.RemoveAliases(data.New.RemovedAliases); err != nil {
			return nil, err
		}
		if err := qb.UpdateAliases(updatedPerformer.ID, currentAliases); err != nil {
			return nil, err
		}

		return updatedPerformer, nil
	default:
		return nil, errors.New("Unsupported operation: " + operation.String())
	}
}
