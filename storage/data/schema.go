package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const schema1 = `

CREATE TABLE [cameras] (
    [ID] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    [Name] VARCHAR UNIQUE,
    [Type] VARCHAR,
    [Host] VARCHAR UNIQUE,
    [Enabled] BOOLEAN DEFAULT TRUE,
    [LastSeen] DATETIME);

CREATE TABLE [files] (
    [ID] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    [Path] VARCHAR NOT NULL UNIQUE,
    [Type] TINYINT NOT NULL,
    [Timestamp] DATETIME NOT NULL,
    [Received] DATETIME,
    [CameraID] INTEGER,
    [DurationSeconds] INTEGER,
    FOREIGN KEY([CameraID]) REFERENCES [cameras]([ID])
);

CREATE  INDEX idx_file_cameraid_timestamp 
ON files (CameraID, TimeStamp);
`

// Rather than dealing with loose files on disk (which just adds complication
// for deployment and config, we put them in as strings in this file,
// then write them out, in order, to disk, then use the migrate file driver
// to pull them back in.  I didn't see a direct driver where I can just pass 
// in files in order, this is simple enough.
func getSchemaDir() (string, func(), error) {

	tempDir := path.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().Unix()))

	done := func(){}

	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return "", done,fmt.Errorf("Error creating schema dir: %v", err)
	}
	
	for i, file := range []string{
								schema1,
	} {
		fileName := path.Join(tempDir, fmt.Sprintf("%d_schema.up.sql", i+1))
		err := ioutil.WriteFile(fileName, []byte(file), os.ModePerm)
		if err != nil {
			return "", done, fmt.Errorf("Failed to write schema files: %v", err)
		}
	}

	// cleanup the temp dir
	done = func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, done, nil
}