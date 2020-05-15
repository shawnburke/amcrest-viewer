package data

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
)

type Repository interface {
	AddCamera(name string, t string, host *string) (*entities.Camera, error)
	GetCamera(id int) (*entities.Camera, error)
	DeleteCamera(id int) (bool, error)
	UpdateCamera(id int, name *string, host *string, enabled *bool) (*entities.Camera, error)
	SeenCamera(id int) error
}

func NewRepository(db *sqlx.DB) (Repository, error) {
	return &sqlRepository{
		db: db,
	}, nil
}

type sqlRepository struct {
	db *sqlx.DB
}

func (sr *sqlRepository) AddCamera(name string, t string, host *string) (*entities.Camera, error) {

	result, err := sr.db.Exec(
		`INSERT INTO cameras 
		(Name, Type)
		VALUES
		($1,$2)`, name, t)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if host != nil {
		_, err := sr.db.Exec(`
		UPDATE cameras 
		SET Host=$1 
		WHERE ID=$2`, *host, id)
		if err != nil {
			return nil, err
		}
	}

	return sr.GetCamera(int(id))
}

func (sr *sqlRepository) GetCamera(id int) (*entities.Camera, error) {
	result, err := sr.db.Queryx(`SELECT * FROM cameras WHERE ID=$1`, id)

	if err != nil {
		return nil, fmt.Errorf("Error fetching cam %d: %v", id, err)
	}

	defer result.Close()

	for result.Next() {
		cam := &entities.Camera{}
		if err = result.StructScan(cam); err != nil {
			return nil, err
		}
		return cam, nil
	}

	return nil, os.ErrNotExist
}

func (sr *sqlRepository) DeleteCamera(id int) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (sr *sqlRepository) UpdateCamera(id int, name *string, host *string, enabled *bool) (*entities.Camera, error) {
	panic("not implemented") // TODO: Implement
}

func (sr *sqlRepository) SeenCamera(id int) error {
	panic("not implemented") // TODO: Implement
}
