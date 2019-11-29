package models

import (
    "github.com/satori/go.uuid"

	"github.com/stashapp/stashdb/pkg/database"
)

const (
	studioTable   = "studios"
	studioJoinKey = "studio_id"
)

var (
	studioDBTable = database.NewTable(studioTable, func() interface{} {
		return &Studio{}
	})

	studioUrlTable = database.NewTableJoin(studioTable, "studio_urls", studioJoinKey, func() interface{} {
		return &StudioUrl{}
	})
)

type Studio struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	Name           string          `db:"name" json:"name"`
    Image          []byte          `db:"image" json:"image"`
	ParentStudioID uuid.NullUUID   `db:"parent_studio_id,omitempty" json:"parent_studio_id"`
	CreatedAt      SQLiteTimestamp `db:"created_at" json:"created_at"`
	UpdatedAt      SQLiteTimestamp `db:"updated_at" json:"updated_at"`
}

func (Studio) GetTable() database.Table {
	return studioDBTable
}

func (p Studio) GetID() uuid.UUID {
	return p.ID
}

type Studios []*Studio

func (p Studios) Each(fn func(interface{})) {
	for _, v := range p {
		fn(v)
	}
}

func (p *Studios) Add(o interface{}) {
	*p = append(*p, o.(*Studio))
}

type StudioUrl struct {
	StudioID uuid.UUID `db:"studio_id" json:"studio_id"`
	URL      string    `db:"url" json:"url"`
	Type     string    `db:"type" json:"type"`
}

func (p *StudioUrl) ToURL() URL {
	return URL{
		URL:  p.URL,
		Type: p.Type,
	}
}

type StudioUrls []StudioUrl

func (p StudioUrls) Each(fn func(interface{})) {
	for _, v := range p {
		fn(v)
	}
}

func (p *StudioUrls) Add(o interface{}) {
	*p = append(*p, o.(StudioUrl))
}

func CreateStudioUrls(studioId uuid.UUID, urls []*URLInput) []StudioUrl {
	var ret []StudioUrl

	for _, urlInput := range urls {
		ret = append(ret, StudioUrl{
			StudioID: studioId,
			URL:      urlInput.URL,
			Type:     urlInput.Type,
		})
	}

	return ret
}

func (p *Studio) IsEditTarget() {
}

func (p *Studio) CopyFromCreateInput(input StudioCreateInput) {
	CopyFull(p, input)

    if input.ParentID != nil {
        UUID, err := uuid.FromString(*input.ParentID)
        if err == nil {
            p.ParentStudioID = uuid.NullUUID{UUID: UUID, Valid: true}
        }
    }
}

func (p *Studio) CopyFromUpdateInput(input StudioUpdateInput) {
	CopyFull(p, input)

    if input.ParentID != nil {
        UUID, err := uuid.FromString(*input.ParentID)
        if err == nil {
            p.ParentStudioID = uuid.NullUUID{UUID: UUID, Valid: true}
        }
    }
}
