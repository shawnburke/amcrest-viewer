package scanner

import (
	"testing"
	"time"

	"encoding/json"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTimeStamps(t *testing.T) {

	p := "2019-05-10/001/dav/12/12.52.24-12.52.48[M][0@0][0].mp4"

	ts := pathToTimestamps(p)

	require.Len(t, ts, 2)

	require.Equal(t, "2019-05-10 12:52:24 -0700 PDT", ts[0].String())

}

func TestJpgTimeStamps(t *testing.T) {
	p := "2019-05-10/001/jpg/02/29/41[M][0@0][0].jpg"

	ts := jpgPathToTimestamp(p)
	exp, _ := time.Parse(time.RFC3339, "2019-05-10T02:29:41.000Z")
	require.Equal(t, exp, ts)
}

var vidJSON = `{
  "Time": "2019-05-10T00:05:35-07:00",
  "Path": "2019-05-10/001/dav/00/00.05.35-00.06.00[M][0@0][0].mp4",
  "Duration": 25000000000,
  "Images": [
    {
      "Time": "2019-05-10T00:05:42-07:00",
      "Path": "2019-05-10/001/jpg/00/05/42[M][0@0][0].jpg"
    },
    {
      "Time": "2019-05-10T00:05:43-07:00",
      "Path": "2019-05-10/001/jpg/00/05/43[M][0@0][0].jpg"
    },
    {
      "Time": "2019-05-10T00:05:44-07:00",
      "Path": "2019-05-10/001/jpg/00/05/44[M][0@0][0].jpg"
    }
  ]
}`

func TestWalk(t *testing.T) {

	files, err := FindFiles("./test_data", zap.NewNop(), false)

	require.NoError(t, err)

	for _, f := range files {
		if f.DateString() == "2019-05-10" {
			vid := f.Videos[0]
			require.NotNil(t, vid)

			raw, _ := json.MarshalIndent(vid, "", "  ")

			require.Equal(t, vidJSON, string(raw))
			return

		}
	}

}
