import React from 'react';
import { Card } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { Performers_queryPerformers_performers as Performer } from 'src/definitions/Performers';

import { getUrlByType } from 'src/utils/transforms';

interface PerformerCardProps {
    performer: Performer;
}

const CLASSNAME = 'PerformerCard';
const CLASSNAME_IMAGE = `${CLASSNAME}-image`;

const PerformerCard: React.FC<PerformerCardProps> = ({ performer }) => (
    <div key={performer.id} className={`col-12 col-lg-3 col-md-6 ${CLASSNAME}`}>
        <Card>
            <Link to={`/performers/${performer.id}`}>
                <div className={CLASSNAME_IMAGE}>
                    <img src={getUrlByType(performer.urls, 'PHOTO', 'portrait')} alt="" />
                </div>
                <Card.Footer><h5>{performer.name}</h5></Card.Footer>
            </Link>
        </Card>
    </div>
);

export default PerformerCard;