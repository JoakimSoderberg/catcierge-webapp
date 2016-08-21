package main

import (
	"time"
	"labix.org/v2/mgo/bson"
)

type CatEventData struct {
    ID                  string `json:"id"`
    EventJSONVersion    string `json:"event_json_version"`
    CatciergeType       string `json:"catcierge_type"`
    Description         string `json:"description"`
    Start               time.Time `json:"start"`
    End                 string `json:"end"`
    TimeGenerated       time.Time `json:"time_generated"`
    Timezone            string `json:"timezone"`
    TimezoneUtcOffset   string `json:"timezone_utc_offset"`
    GitHash             string `json:"git_hash"`
    GitHashShort        string `json:"git_hash_short"`
    GitTainted          int    `json:"git_tainted"`
    MatchGroupCount     int    `json:"match_group_count"`
    MatchGroupDirection string `json:"match_group_direction"`
    MatchGroupMaxCount  int    `json:"match_group_max_count"`
    MatchGroupSuccess   int    `json:"match_group_success"`
    Rootpath            string `json:"rootpath"`
    State               string `json:"state"`
    PrevState           string `json:"prev_state"`
    Version             string `json:"version"`
    Matches             []struct {
        ID          string `json:"id"`
        Description string `json:"description"`
        Directon    string `json:"direction"`
        Filename    string `json:"filename"`
        Path        string `json:"path"`
        Result      int    `json:"result"`
        Success     int    `json:"success"`
        Time        time.Time `json:"time"`
        IsFalsePositive bool `json:"is_false_positive"`
        StepCount   int    `json:"step_count"`
        Steps       []struct {
            Active      int    `json:"active"`
            Description string `json:"description"`
            Filename    string `json:"filename"`
            Name        string `json:"name"`
            Path        string `json:"path"`
        } `json:"steps"`
    } `json:"matches"`
    Settings  struct {
        HaarMatcher struct {
            Cascade       string `json:"cascade"`
            EqHistogram   int    `json:"eq_histogram"`
            InDirection   string `json:"in_direction"`
            MinSizeHeight int    `json:"min_size_height"`
            MinSizeWidth  int    `json:"min_size_width"`
            NoMatchIsFail int    `json:"no_match_is_fail"`
            PreyMethod    string `json:"prey_method"`
            PreySteps     int    `json:"prey_steps"`
        } `json:"haar_matcher"`
        LockoutError      int    `json:"lockout_error"`
        LockoutErrorDelay int    `json:"lockout_error_delay"`
        LockoutMethod     int    `json:"lockout_method"`
        LockoutTime       int    `json:"lockout_time"`
        Matcher           string `json:"matcher"`
        Matchtime         int    `json:"matchtime"`
        NoFinalDecision   int    `json:"no_final_decision"`
        OkMatchesNeeded   int    `json:"ok_matches_needed"`
    } `json:"settings"`
}

type CatEvent struct {
    ID bson.ObjectId    `json:"id", bson:"_id"`
    Name string			`json:"name"`
    Data CatEventData   `json:"data", bson:"data"`
    Tags []string       `json:"tags", bson:"tags"`
}

type CatEventResource struct {
    // TODO: Replace with MongoDB
    events map[string]CatEvent
}