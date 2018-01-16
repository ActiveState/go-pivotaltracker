package pivotal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Activity is ...
type Activity struct {
	Kind             string        `json:"kind"`
	GUID             string        `json:"guid"`
	ProjectVersion   int           `json:"project_version"`
	Message          string        `json:"message"`
	Highlight        string        `json:"highlight"`
	Changes          []interface{} `json:"changes"`
	PrimaryResources []interface{} `json:"primary_resources"`
	Project          interface{}   `json:"project"`
	PerformedBy      interface{}   `json:"performed_by"`
	OccurredAt       time.Time     `json:"occurred_at"`
}

// ActivityService is ...
type ActivityService struct {
	client *Client
}

func newActivitiesService(client *Client) *ActivityService {
	return &ActivityService{client}
}

// List returns all activities matching the filter in case the filter is specified.
//
// List actually sends 2 HTTP requests - one to get the total number of activities,
// another to retrieve the activities using the right pagination setup. The reason
// for this is that the filter might require to fetch all the activities at once
// to get the right results. The response is default sorted in DESCENDING order so 
// leverage the sortAsc variable to control sort order.
func (service *ActivityService) List(projectID int, version int, sortAsc bool) ([]*Activity, error) {
	reqFunc := newActivitiesRequestFunc(service.client, projectID, version, sortAsc)
	cursor, err := newCursor(service.client, reqFunc, 0)
	if err != nil {
		return nil, err
	}

	var activities []*Activity
	if err := cursor.all(&activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func newActivitiesRequestFunc(client *Client, projectID int, version int, sortAsc bool) func() *http.Request {
	return func() *http.Request {
		v := strconv.Itoa(version)
		u := fmt.Sprintf("projects/%v/activity", projectID)
		if v != "" {
			u += "?since_version=" + url.QueryEscape(v)
		}
		u += "&sort_order="
		if sortAsc {
			u += url.QueryEscape("asc")
		} else {
			u += url.QueryEscape("desc")
		}
		req, _ := client.NewRequest("GET", u, nil)
		return req
	}
}

// ActivityCursor is...
type ActivityCursor struct {
	*cursor
	buff []*Activity
}

// Next returns the next story.
//
// In case there are no more stories, io.EOF is returned as an error.
func (c *ActivityCursor) Next() (s *Activity, err error) {
	if len(c.buff) == 0 {
		_, err = c.next(&c.buff)
		if err != nil {
			return nil, err
		}
	}

	if len(c.buff) == 0 {
		err = io.EOF
	} else {
		s, c.buff = c.buff[0], c.buff[1:]
	}
	return s, err
}

// Iterate returns a cursor that can be used to iterate over the activities specified
// by the filter. More stories are fetched on demand as needed.
func (service *ActivityService) Iterate(projectID int, version int, sortAsc bool) (c *ActivityCursor, err error) {
	reqFunc := newActivitiesRequestFunc(service.client, projectID, version, sortAsc)
	cursor, err := newCursor(service.client, reqFunc, PageLimit)
	if err != nil {
		return nil, err
	}
	return &ActivityCursor{cursor, make([]*Activity, 0)}, nil
}
