package eol

import (
	"encoding/json"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"gopkg.in/tomb.v2"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"net/url"
)

type SearchResult struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Content string `json:"content"`
}

type SearchQuery struct {
	// Query is the string to search for.
	Query string
	// Exact will find taxon pages if the title or any synonym or common name exactly matches the search term.
	Exact bool
	// Limit will restrict the number of results to the integer specified.
	Limit int
	// FilterByTaxonConceptID limits search results to members of that EOL page ID taxonomic group.
	FilterByTaxonConceptID int
	// FilterByHierarchyEntryID limits search results to members of that taxonomic group.
	FilterByHierarchyEntryID int
	// FilterByString, when provided, will make an exact search and that matching page will be used as the taxonomic group against which to filter search results.
	FilterByString string
	// CacheTTL specifies the number of seconds you wish to have the response cached.
	CacheTTL int
	////////////////////////////////
	tmb tomb.Tomb
	ch  chan SearchResult
}

func Search(q SearchQuery) (response []SearchResult, err error) {


	if q.Query == "" {
		return nil, errors.New("a query value is required for eol search")
	}

	for r := range q.next() {
		response = append(response, r)
	}

	if err := q.tmb.Err(); err != nil {
		return nil, err
	}

	return

}

func (s *SearchQuery) next() <-chan SearchResult {
	s.tmb = tomb.Tomb{}
	s.ch = make(chan SearchResult, 5)
	go s.tmb.Go(func() error {
		err := s.request(1)
		close(s.ch)
		return err
	})
	return s.ch
}

type results struct {
	TotalResults float64        `json:"totalResults"`
	ItemsPerPage float64        `json:"itemsPerPage"`
	Results      []SearchResult `json:"results"`
}

func (s *SearchQuery) request(page int) error {

	select {
	case <-s.tmb.Dying():
		return nil
	default:
	}

	resp, err := http.Get(s.url(page))
	if err != nil {
		return errors.Wrap(err, "could not get http response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Wrapf(errors.New(resp.Status), "StatusCode: %d; URL: %s", resp.StatusCode, s.url(page))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read http response body")
	}
	var response results
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "could not unmarshal http response")
	}

	for i := range response.Results {
		s.ch <- response.Results[i]
	}
	if page > 1 {
		return nil
	}

	// If this the first page, schedule the remaining requests
	totalRequests := math.Ceil(response.TotalResults / response.ItemsPerPage)
	var wg sync.WaitGroup
	for i := 2; i <= int(totalRequests); i++ {
		func(count int) {
			wg.Add(1)
			go s.tmb.Go(func() error {
				defer wg.Done()
				return s.request(count)
			})
		}(i)
	}
	wg.Wait()
	return nil
}

func (q *SearchQuery) url(page int) string {

	str := fmt.Sprintf(
		"http://eol.org/api/search/1.0.json?q=%s&exact=%v&filter_by_string=%s&cache_ttl=%d&page=%d",
		url.QueryEscape(q.Query),
		q.Exact,
		q.FilterByString,
		q.CacheTTL,
		page,
	)

	if q.FilterByHierarchyEntryID > 0 {
		str += fmt.Sprintf("filter_by_hierarchy_entry_id=%d", q.FilterByHierarchyEntryID)
	}

	if q.FilterByTaxonConceptID > 0 {
		str += fmt.Sprintf("filter_by_taxon_concept_id=%d", q.FilterByTaxonConceptID)
	}

	if q.Limit > 0 {
		str += fmt.Sprintf("limit=%d", q.Limit)
	}

	return str
}
