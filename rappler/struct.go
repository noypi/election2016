package election2016

type JsData struct {
	Timestamp    int64
	Contest_data []*JsBallotData
}

type JsBallotData struct {
	Ballot_name string
	Persson_id  int
	Values      []*JsBallotHistory
}

type JsBallotHistory struct {
	Person_id            int
	Last_name            string
	First_name           string
	Slug                 string
	Political_party_name string
	Vote_count           int64
	Date                 int64
}

type JsLoc struct {
	Location_id             int
	Location_name           string
	Timestamp               int64
	Total_registered_voters int64
}
