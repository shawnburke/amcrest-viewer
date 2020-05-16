package data

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
)


type Repository interface {

	// Camera operations
	AddCamera(name string, t string, host *string) (*entities.Camera, error)
	GetCamera(id int) (*entities.Camera, error)
	DeleteCamera(id int) (bool, error)
	UpdateCamera(id int, name *string, host *string, enabled *bool) (*entities.Camera, error)
	SeenCamera(id int) error
	ListCameras() ([]*entities.Camera, error)

	// File operations
	AddFile(path string, t int, cameraID int, timestamp time.Time, duration *time.Duration) (*entities.File, error)
	GetFile(id int) (*entities.File, error)
	ListFiles(cameraID int, start *time.Time, end *time.Time, fileType *int) ([]*entities.File, error)
}

func NewRepository(db *sqlx.DB, logger *zap.Logger) (Repository, error) {
	return &sqlRepository{
		db: db,
		logger: logger,
	}, nil
}

type sqlRepository struct {
	db *sqlx.DB
	logger *zap.Logger
}

func (sr *sqlRepository) AddCamera(name string, t string, host *string) (*entities.Camera, error) {

	tx, err :=  sr.db.Begin()

	if err != nil {
		return nil, fmt.Errorf("Failed to start txn: %w", err)
	}

	rollback := func() {
		
		if rbErr := tx.Rollback(); rbErr != nil{
			sr.logger.Error("Error canceling transaction", zap.Error(rbErr))
		}
	}

	result, err := sr.db.Exec(
		`INSERT INTO cameras 
		(Name, Type)
		VALUES
		($1,$2)`, name, t)

	if err != nil {
		rollback()
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		rollback()
		return nil, err
	}

	if host != nil {
		_, err := sr.db.Exec(`
		UPDATE cameras 
		SET Host=$1 
		WHERE ID=$2`, *host, id)
		if err != nil {
			rollback()
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	
	return sr.GetCamera(int(id))
}

func (sr *sqlRepository) GetCamera(id int) (*entities.Camera, error) {
	result, err := sr.db.Queryx(`SELECT * FROM cameras WHERE ID=$1`, id)

	if err != nil {
		return nil, fmt.Errorf("Error fetching cam %d: %w", id, err)
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

func (sr *sqlRepository) 	ListCameras() ([]*entities.Camera, error){

	result, err := sr.db.Queryx(`SELECT * FROM cameras`)

	if err != nil {
		return nil, fmt.Errorf("Error fetching cams: %w", err)
	}

	defer result.Close()

	cams := make([]*entities.Camera, 0, 8)
	for result.Next() {
		cam := &entities.Camera{}
		if err = result.StructScan(cam); err != nil {
			return nil, err
		}
		cams = append(cams, cam)
	}

	return cams,nil

}

func (sr *sqlRepository) DeleteCamera(id int) (bool, error) {
	result, err := sr.db.Exec(`DELETE FROM cameras WHERE ID=$1`, id)

	if err != nil {
		return false, fmt.Errorf("Error deleting camera %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if rows > 0 {
		if err != nil {
			sr.logger.Error("Error getting Rows affected", zap.Error(err))
		}
		return true, nil
	}

	return false, fmt.Errorf("Error deleting camera %d: %w", id, err)
}

func (sr *sqlRepository) UpdateCamera(id int, name *string, host *string, enabled *bool) (*entities.Camera, error) {
	
	tx, err := sr.db.Begin()

	update := func(field string, value interface{}) error {
		_, err := sr.db.Exec(fmt.Sprintf("UPDATE cameras SET %s = $1", field), value)
		if err != nil {
			return fmt.Errorf("Error updating field %s: %w", field, err)
		}
		return nil
	}

	updates := []struct{
		Field string
		Value interface{}
	}{
		{
			Field: "Name",
			Value: name,
		},
		{
			Field: "Host",
			Value: host,
		},
		{
			Field: "Enabled",
			Value: enabled,
		},
	}

	for _, u := range updates {
		if u.Value == nil || reflect.ValueOf(u.Value).IsNil(){
			continue
		}

		err := update(u.Field, u.Value)

		if err != nil {
			if e2 := tx.Rollback(); e2 != nil {
				sr.logger.Error("Failed to rollback", zap.Error(e2))
			}
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil  {
		return nil, fmt.Errorf("Faield to commit: %w", err)
	}

	return sr.GetCamera(id)

}

func (sr *sqlRepository) SeenCamera(id int) error {
	 _, err := sr.db.Exec(`UPDATE cameras SET LastSeen=$1 WHERE ID=$2`, time.Now(), id)
	 return err
}

//  Files

func (sr *sqlRepository) AddFile(path string, t int, cameraID int, timestamp time.Time, duration *time.Duration) (*entities.File, error) {
	
	query := `
	INSERT INTO files
		(Path, Type, CameraID, Timestamp)
		VALUES
		($1,$2,$3,$4)
	`

	args := []interface{}{
		path,
		t,
		cameraID,
		timestamp,
	}

	if duration != nil {
query = `
	INSERT INTO files
		(Path, Type, CameraID, Timestamp, DurationSeconds)
		VALUES
		($1,$2,$3,$4,$5)
	`
	args = append(args, int(duration.Seconds()))

	}

	
	res, err := sr.db.Exec(query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to add file: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("Failed to get id: %w", err)
	}
	return sr.GetFile(int(id))
}

func (sr *sqlRepository) GetFile(id int) (*entities.File,error) {
	result, err := sr.db.Queryx(`SELECT * FROM files WHERE ID=$1`, id)

	if err != nil {
		return nil, fmt.Errorf("Error fetching file %d: %w", id, err)
	}

	defer result.Close()

	for result.Next() {
		f := &entities.File{}
		if err = result.StructScan(f); err != nil {
			return nil, err
		}
		return f, nil
	}

	return nil, os.ErrNotExist
}
func (sr *sqlRepository) ListFiles(cameraID int, start *time.Time, end *time.Time, fileType *int) ([]*entities.File, error) {
	query := `SELECT * FROM files WHERE CameraID=$1 `


	args := []interface{}{
		cameraID,
	}

	if start != nil {
		query += " AND Timestamp >= $2"
		args = append(args, *start)

		if end != nil {
			query += " AND Timestamp < $3"
			args = append(args, *end)
		}
	}

	if fileType != nil {
		query += fmt.Sprintf("AND [Type]=$%d", len(args)+1)
		args = append(args, *fileType)
	}

	result, err := sr.db.Queryx(query, args...)

	if err != nil {
		return nil, fmt.Errorf("Error fetching files: %w", err)
	}

	defer result.Close()

	files := make([]*entities.File, 0, 8)
	for result.Next() {
		file := &entities.File{}
		if err = result.StructScan(file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files,nil
}