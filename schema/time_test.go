package schema_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rs/rest-layer/schema"
)

var testTimeFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
}

// timeValidateTest can be used for positive validation tests only; negative
// tests should be hand-written to check for the correct errors.
type timeValidateTest struct {
	validator  schema.Time
	input      interface{}
	expectTime time.Time
}

func (tt timeValidateTest) Run(t *testing.T) {
	t.Parallel()

	v := &tt.validator
	v.Compile(nil)
	value, err := v.Validate(tt.input)

	t.Run("should not error", func(t *testing.T) {
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("should return the expected value", func(t *testing.T) {
		ts, ok := value.(time.Time)
		if !ok {
			t.Errorf("expected type time.Time, got type %T", value)
		}
		if !ts.Equal(tt.expectTime) {
			t.Errorf("expected time %s, got %s", tt.expectTime, ts)
		}
		tsLoc := ts.Location().String()
		expectLoc := tt.expectTime.Location().String()

		if expectLoc != tsLoc {
			t.Errorf("expected time-zone %v, got %v", expectLoc, tsLoc)
		}
	})
}

func TestTimeValidate(t *testing.T) {
	tsString := "2018-11-18T17:15:16.000000017Z"
	tsParsed, _ := time.Parse(time.RFC3339Nano, tsString)
	tzMinus1 := time.FixedZone("UTC-1,", -60*60)

	t.Run("when validating a time string", timeValidateTest{
		validator:  schema.Time{},
		input:      tsString,
		expectTime: time.Date(2018, 11, 18, 17, 15, 16, 17, time.UTC),
	}.Run)

	t.Run("when validating a parsed time", timeValidateTest{
		validator:  schema.Time{},
		input:      tsParsed,
		expectTime: time.Date(2018, 11, 18, 17, 15, 16, 17, time.UTC),
	}.Run)
	t.Run("when changing time-zone", timeValidateTest{
		validator: schema.Time{
			Location: tzMinus1,
		},
		input:      tsParsed,
		expectTime: time.Date(2018, 11, 18, 16, 15, 16, 17, tzMinus1),
	}.Run)
	t.Run("when truncating string to one seconde", timeValidateTest{
		validator: schema.Time{
			Truncate: time.Second,
		},
		input:      tsParsed,
		expectTime: time.Date(2018, 11, 18, 17, 15, 16, 0, time.UTC),
	}.Run)

	t.Run("when truncating a parsed time to 24 hours", timeValidateTest{
		validator: schema.Time{
			Truncate: time.Hour * 24,
		},
		input:      tsParsed,
		expectTime: time.Date(2018, 11, 18, 0, 0, 0, 0, time.UTC),
	}.Run)

	t.Run("when truncating to 24 hours and changing time-zone", timeValidateTest{
		validator: schema.Time{
			Truncate: time.Hour * 24,
			Location: tzMinus1,
		},
		input:      tsParsed,
		expectTime: time.Date(2018, 11, 17, 23, 0, 0, 0, tzMinus1),
	}.Run)
}
func TestTimeSpecificLayoutList(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	// list to test for
	testList := []string{time.RFC1123Z, time.RFC822Z, time.RFC3339}
	// test for same list in reverse
	timeT := schema.Time{TimeLayouts: []string{time.RFC3339, time.RFC822Z, time.RFC1123Z}}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	// expect no errors
	for _, f := range testList {
		_, err := timeT.Validate(now.Format(f))
		assert.NoError(t, err)
	}
}

func TestTimeForTimeLayoutFailure(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	// test for ANSIC time
	testList := []string{time.ANSIC}
	// configure for RFC3339 time
	timeT := schema.Time{TimeLayouts: []string{time.RFC3339}}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	// expect an error
	for _, f := range testList {
		_, err := timeT.Validate(now.Format(f))
		assert.EqualError(t, err, "not a time")
	}
}

func TestTimeLess(t *testing.T) {
	low, _ := time.Parse(time.RFC3339, "2018-11-18T17:15:16Z")
	high, _ := time.Parse(time.RFC3339, "2018-11-19T17:15:16Z")
	cases := []struct {
		name         string
		value, other interface{}
		expected     bool
	}{
		{`Time.Less(time.Time-low,time.Time-high)`, low, high, true},
		{`Time.Less(time.Time-low,time.Time-low)`, low, low, false},
		{`Time.Less(time.Time-high,time.Time-low)`, high, low, false},
		{`Time.Less(time.Time,string)`, low, "2.0", false},
	}
	lessFunc := schema.Time{}.LessFunc()
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := lessFunc(tt.value, tt.other)
			if got != tt.expected {
				t.Errorf("output for `%v`\ngot:  %v\nwant: %v", tt.name, got, tt.expected)
			}
		})
	}
}
