package data

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	cameras "github.com/shawnburke/amcrest-viewer/cameras/common"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"go.uber.org/zap"
)

type Repository interface {

	// Camera operations
	AddCamera(name string, t string, host *string) (*entities.Camera, error)
	GetCamera(id string) (*entities.Camera, error)
	GetCameraStats(id string, start *time.Time, end *time.Time, breakdown string) (*CameraStats, error)
	DeleteCamera(id string) (bool, error)
	UpdateCamera(id string, name *string, host *string, enabled *bool) (*entities.Camera, error)
	UpdateCameraCreds(id string, ip string, username string, password string) error
	SeenCamera(id string) error
	ListCameras() ([]*entities.Camera, error)

	// File operations
	AddFile(path string, t int, cameraID string, length int, timestamp time.Time, duration *time.Duration) (*entities.File, error)
	DeleteFile(id int) error
	GetFile(id int) (*entities.File, error)

	ListFiles(cameraID string, filter *ListFilesFilter) ([]*entities.File, error)
}

type ListFilesFilter struct {
	Start      *time.Time
	End        *time.Time
	FileType   *int
	Descending bool
}

const (
	StatFlagBreakdown = 0x00000004
)

type CameraStats struct {
	FileCount  int              `json:"file_count"`
	FileCounts map[int]FileData `json:"file_counts"`
	FileSize   int              `json:"file_size"`
	MinDate    time.Time        `json:"min_date"`
	MaxDate    time.Time        `json:"max_date"`
	Breakdown  []FileData       `json:"breakdown,omitempty"`
}

type FileData struct {
	Count int `json:"count"`
	Size  int `json:"size"`
}

func NewRepository(db *sqlx.DB, t common.Time, logger *zap.Logger, bus common.EventBus) (Repository, error) {
	return &sqlRepository{
		db:     db,
		time:   t,
		logger: logger,
		bus:    bus,
	}, nil
}

type sqlRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
	time   common.Time
	bus    common.EventBus
}

func (sr *sqlRepository) AddCamera(name string, t string, host *string) (*entities.Camera, error) {

	if name == "" {
		return nil, fmt.Errorf("Name required")
	}

	if t == "" {
		return nil, fmt.Errorf("Type required")
	}

	tx, err := sr.db.Begin()

	if err != nil {
		return nil, fmt.Errorf("Failed to start txn: %w", err)
	}

	rollback := func() {

		if rbErr := tx.Rollback(); rbErr != nil {
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

	cam := entities.Camera{
		ID:   int(id),
		Type: t,
	}

	return sr.GetCamera(cam.CameraID())
}

var camIDRegEx = regexp.MustCompile(`([^-]+)-(\d+)`)

func parseCameraID(cameraID string) (int, error) {
	match := camIDRegEx.FindStringSubmatch(cameraID)

	if match == nil {
		return -1, fmt.Errorf("Invalid cameraID: %s", cameraID)
	}

	id, err := strconv.Atoi(match[2])
	if err != nil {
		return -1, err
	}

	return id, err
}

func (sr *sqlRepository) GetCamera(cameraID string) (*entities.Camera, error) {

	id, err := parseCameraID(cameraID)
	if err != nil {
		return nil, err
	}

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

func (sr *sqlRepository) getTimeRange(start *time.Time, end *time.Time) (time.Time, time.Time) {
	s := time.Time{}
	e := sr.time.Now().AddDate(100, 0, 0)

	if start != nil {
		s = *start
	}

	if end != nil {
		e = *end
	}
	return s.UTC(), e.UTC()
}

func (sr *sqlRepository) GetCameraStats(cameraID string, start *time.Time, end *time.Time, breakdown string) (*CameraStats, error) {

	id, err := parseCameraID(cameraID)
	if err != nil {
		return nil, err
	}

	var query = `
			SELECT [Type], COUNT(Id) as TotalCount, SUM(Length) as TotalSize, datetime(MIN(Timestamp)) as Min, datetime(MAX(Timestamp)) as Max  
			FROM files WHERE 
			CameraId=$1 AND (Timestamp >= $2 AND Timestamp < $3) GROUP BY [Type]`

	s, e := sr.getTimeRange(start, end)

	res, err := sr.db.Queryx(query, id, s, e)

	if err != nil {
		return nil, fmt.Errorf("Error getting camera stats: %w", err)
	}
	defer res.Close()

	cs := &CameraStats{
		FileCounts: map[int]FileData{},
	}

	for res.Next() {
		r := struct {
			Type       int    `db:"Type"`
			TotalCount int    `db:"TotalCount"`
			TotalSize  int    `db:"TotalSize"`
			Min        string `db:"Min"`
			Max        string `db:"Max"`
		}{}

		err = res.StructScan(&r)
		if err != nil {
			return nil, err
		}

		cs.FileCount += r.TotalCount
		cs.FileSize += r.TotalSize

		dateTimeFormat := "2006-01-02 15:04:05"

		min, err := time.Parse(dateTimeFormat, r.Min)
		if err != nil {
			return nil, err
		}
		max, err := time.Parse(dateTimeFormat, r.Max)
		if err != nil {
			return nil, err
		}
		if cs.MinDate.IsZero() || cs.MinDate.After(min) {
			cs.MinDate = min
		}

		if cs.MaxDate.IsZero() || cs.MaxDate.Before(max) {
			cs.MaxDate = max
		}

		cs.FileCounts[r.Type] = FileData{
			Size:  r.TotalSize,
			Count: r.TotalCount,
		}

	}
	return cs, nil

}

func (sr *sqlRepository) ListCameras() ([]*entities.Camera, error) {

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

	return cams, nil

}

func (sr *sqlRepository) DeleteCamera(cameraID string) (bool, error) {

	id, err := parseCameraID(cameraID)
	if err != nil {
		return false, err
	}

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

func (sr *sqlRepository) UpdateCamera(cameraID string, name *string, host *string, enabled *bool) (*entities.Camera, error) {

	camID, err := parseCameraID(cameraID)
	if err != nil {
		return nil, err
	}

	tx, err := sr.db.Begin()

	update := func(field string, value interface{}) error {
		_, err := sr.db.Exec(fmt.Sprintf("UPDATE cameras SET %s = $1 WHERE ID=$2", field), value, camID)
		if err != nil {
			return fmt.Errorf("Error updating field %s: %w", field, err)
		}
		return nil
	}

	updates := []struct {
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
		if u.Value == nil || reflect.ValueOf(u.Value).IsNil() {
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

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("Faield to commit: %w", err)
	}

	return sr.GetCamera(cameraID)
}

func (sr *sqlRepository) UpdateCameraCreds(cameraID string, host, user, pass string) error {

	camID, err := parseCameraID(cameraID)
	if err != nil {
		return err
	}

	_, err = sr.db.Exec("UPDATE cameras SET Host=$1,Username=$2,Password=$3 WHERE ID=$r", host, user, pass, camID)

	if err == nil && sr.bus != nil {
		// notify that camera creds have changed.
		event := cameras.NewCameraCredsChangeEvent(cameraID)
		sr.bus.Send(event)
	}
	return err
}

func (sr *sqlRepository) SeenCamera(cameraID string) error {

	id, err := parseCameraID(cameraID)
	if err != nil {
		return err
	}
	_, err = sr.db.Exec(`UPDATE cameras SET LastSeen=$1 WHERE ID=$2`, sr.time.Now(), id)
	return err
}

//  Files

func (sr *sqlRepository) AddFile(
	path string,
	t int,
	cameraID string,
	length int,
	timestamp time.Time,
	duration *time.Duration,
) (*entities.File, error) {

	camID, err := parseCameraID(cameraID)
	if err != nil {
		return nil, err
	}

	query := `
	INSERT INTO files
		(Path, Type, CameraID, Timestamp, Length)
		VALUES
		($1,$2,$3,$4,$5)
	`

	args := []interface{}{
		path,
		t,
		camID,
		timestamp,
		length,
	}

	if duration != nil {
		query = `
	INSERT INTO files
		(Path, Type, CameraID, Timestamp, Length, DurationSeconds)
		VALUES
		($1,$2,$3,$4,$5,$6)
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

func (sr *sqlRepository) DeleteFile(id int) error {
	_, err := sr.db.Queryx(`DELETE * FROM files WHERE ID=$1`, id)
	return err
}

func (sr *sqlRepository) GetFile(id int) (*entities.File, error) {
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

func (sr *sqlRepository) ListFiles(cameraID string, filter *ListFilesFilter) ([]*entities.File, error) {

	camID, err := parseCameraID(cameraID)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = &ListFilesFilter{}
	}

	query := `SELECT * FROM files 
		WHERE CameraID=$1 AND (Timestamp >= $2 AND Timestamp < $3) `

	s, e := sr.getTimeRange(filter.Start, filter.End)

	args := []interface{}{
		camID,
		s,
		e,
	}

	if filter.FileType != nil {
		query += fmt.Sprintf("AND [Type]=$%d", len(args)+1)
		args = append(args, *filter.FileType)
	}

	if filter.Descending {
		query += " ORDER BY [Timestamp] DESC"
	} else {
		query += " ORDER BY [Timestamp] ASC"
	}

	result, err := sr.db.Queryx(query, args...)

	if err != nil {
		return nil, fmt.Errorf("Error fetching files: %w", err)
	}

	sr.logger.Info("ListFiles range", zap.Any("start", s), zap.Any("end", e))

	defer result.Close()

	files := make([]*entities.File, 0, 8)
	for result.Next() {
		file := &entities.File{}
		if err = result.StructScan(file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}
