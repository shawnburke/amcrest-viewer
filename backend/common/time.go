package common

import "time"

const SnapshotFrequency = time.Minute

type Time interface {
	Now() time.Time
}

type defaultTime struct {
}

func (dt *defaultTime) Now() time.Time {
	return time.Now()
}

func NewTime() Time {
	return &defaultTime{}
}

type testTime struct {
	create time.Time
	track  bool
	base   time.Time
}

func NewTestTime(baseTime time.Time, track bool) Time {
	return &testTime{
		base:   baseTime,
		create: time.Now(),
		track:  track,
	}
}

func (mt *testTime) Now() time.Time {
	t := mt.base
	if mt.track {
		delta := time.Now().Sub(mt.create)
		t = t.Add(delta)
	}
	return t
}
