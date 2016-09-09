package main

import "time"

type CatEventTimeV1 struct {
	time.Time
}

// CatEventTimeV1.UnmarshalJSON Unmarshal a Cat Event Time stamp, the time zone is incorrectly separated
// using a ':' we need to parse it using "2006-01-02T15:04:05.999999999Z0700" instead.
func (c *CatEventTimeV1) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	// Get rid of the quotes "" around the value.
	s = s[1 : len(s)-1]

	t, err := time.Parse(time.RFC3339Nano, s)

	// The first version of the catcierge event JSON uses
	// the wrong format for the timezone without ':' so if we
	// fail to parse we attempt again
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.999999999Z0700", s)
		if err != nil {
			// This will be parsed as UTC which is incorrect...
			// But some dates are in this format.
			t, err = time.Parse("2006-01-02 15:04:05", s)
			return err
		}
	}
	c.Time = t
	return
}

// Header containing the core information such as versions
// and the event ID in the catcierge event JSON.
type CatEventHeader struct {
	ID               string `json:"id" bson:"id"`
	EventJSONVersion string `json:"event_json_version" bson:"event_json_version"`
	Version          string `json:"version" bson:"version"`
	GitHash          string `json:"git_hash" bson:"git_hash"`
	GitHashShort     string `json:"git_hash_short" bson:"git_hash_short"`
	GitTainted       int    `json:"git_tainted" bson:"git_tainted"`
}

// CatEventHaarMatcherSettingsV1 Cat event haar cascade matcher settings.
type CatEventHaarMatcherSettingsV1 struct {
	Cascade       string `json:"cascade" bson:"cascade"`
	EqHistogram   int    `json:"eq_histogram" bson:"eq_histogram"`
	InDirection   string `json:"in_direction" bson:"in_direction"`
	MinSizeHeight int    `json:"min_size_height" bson:"min_size_height"`
	MinSizeWidth  int    `json:"min_size_width" bson:"min_size_width"`
	NoMatchIsFail int    `json:"no_match_is_fail" bson:"no_match_is_fail"`
	PreyMethod    string `json:"prey_method" bson:"prey_method"`
	PreySteps     int    `json:"prey_steps" bson:"prey_steps"`
}

// CatEventSettingsV1 Cat event settings.
type CatEventSettingsV1 struct {
	HaarMatcher       CatEventHaarMatcherSettingsV1 `json:"haar_matcher" bson:"haar_matcher"`
	LockoutError      int                           `json:"lockout_error" bson:"lockout_error"`
	LockoutErrorDelay float32                       `json:"lockout_error_delay" bson:"lockout_error_delay"`
	LockoutMethod     int                           `json:"lockout_method" bson:"lockout_method"`
	LockoutTime       int                           `json:"lockout_time" bson:"lockout_time"`
	Matcher           string                        `json:"matcher" bson:"matcher"`
	Matchtime         int                           `json:"matchtime" bson:"matchtime"`
	NoFinalDecision   int                           `json:"no_final_decision" bson:"no_final_decision"`
	OkMatchesNeeded   int                           `json:"ok_matches_needed" bson:"ok_matches_needed"`
}

// CatEventMatchStepV1 Cat event match step.
type CatEventMatchStepV1 struct {
	Active      int    `json:"active" bson:"active"`
	Description string `json:"description" bson:"description"`
	Filename    string `json:"filename" bson:"filename"`
	Name        string `json:"name" bson:"name"`
	Path        string `json:"path" bson:"path"`
	Ref         string `json:"ref,omitempty"`
}

// CatEventMatchV1 Cat event match.
type CatEventMatchV1 struct {
	ID              string                `json:"id" bson:"id"`
	Description     string                `json:"description" bson:"description"`
	Directon        string                `json:"direction" bson:"direction"`
	Filename        string                `json:"filename" bson:"filename"`
	Path            string                `json:"path" bson:"path"`
	Ref             string                `json:"ref,omitempty"`
	Result          float32               `json:"result" bson:"result"`
	Success         int                   `json:"success" bson:"success"`
	Time            CatEventTimeV1        `json:"time" bson:"time"`
	IsFalsePositive bool                  `json:"is_false_positive" bson:"is_false_positive"`
	StepCount       int                   `json:"step_count" bson:"step_count"`
	Steps           []CatEventMatchStepV1 `json:"steps" bson:"steps"`
}

// CatEventDataV1 Cat event data.
type CatEventDataV1 struct {
	CatEventHeader
	State               string             `json:"state" bson:"state"`
	PrevState           string             `json:"prev_state" bson:"prev_state"`
	CatciergeType       string             `json:"catcierge_type" bson:"catcierge_type"`
	Description         string             `json:"description" bson:"description"`
	Start               CatEventTimeV1     `json:"start" bson:"start"`
	End                 CatEventTimeV1     `json:"end" bson:"end"`
	TimeGenerated       CatEventTimeV1     `json:"time_generated" bson:"time_generated"`
	Timezone            string             `json:"timezone" bson:"timezone"`
	TimezoneUtcOffset   string             `json:"timezone_utc_offset" bson:"timezone_utc_offset"`
	Rootpath            string             `json:"rootpath" bson:"rootpath"`
	MatchGroupCount     int                `json:"match_group_count" bson:"match_group_count"`
	MatchGroupDirection string             `json:"match_group_direction" bson:"match_group_direction"`
	MatchGroupMaxCount  int                `json:"match_group_max_count" bson:"match_group_max_count"`
	MatchGroupSuccess   int                `json:"match_group_success" bson:"match_group_success"`
	Matches             []CatEventMatchV1  `json:"matches" bson:"matches"`
	Settings            CatEventSettingsV1 `json:"settings" bson:"settings"`
}
