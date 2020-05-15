package data

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
`
