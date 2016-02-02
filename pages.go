package eol

import (
"github.com/dropbox/godropbox/errors"
"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type PageQuery struct {
	ID int `json:"id"`
	// limits the number of returned image objects
	Images int `json:"images"`
	// limits the number of returned video objects
	Videos int `json:"videos"`
	// limits the number of returned sound objects
	Sounds int `json:"sounds"`
	// limits the number of returned map objects
	Maps int `json:"maps"`
	// limits the number of returned text objects
	Text int `json:"text"`
	// include the IUCN Red List status object
	IUCN bool `json:"iucn"`
	// 'overview' to return the overview text (if exists), a pipe | delimited list of subject names from the list of EOL accepted subjects (e.g. TaxonBiology, FossilHistory), or 'all' to get text in any subject.
	// Always returns an overview text as a first result (if one exists in the given context).
	Subjects string `json:"subjects"`
	// a pipe | delimited list of licenses or 'all' to get objects under any license.
	// Licenses abbreviated cc- are all Creative Commons licenses.
	// Visit their site for more information on the various licenses they offer.
	// cc-by, cc-by-nc, cc-by-sa, cc-by-nc-sa, pd [public domain], na [not applicable], all
	Licenses string `json:"licenses"`
	// include all metadata for data objects
	Details bool `json:"details"`
	// return all common names for the page's taxon
	CommonNames bool `json:"common_names"`
	// return all synonyms for the page's taxon
	Synonyms bool `json:"synonyms"`
	// return all references for the page's taxon
	References bool `json:"references"`
	// If 'vetted' is given a value of '1', then only trusted content will be returned.
	// If 'vetted' is '2', then only trusted and unreviewed content will be returned (untrusted content will not be returned).
	// The default is to return all content.
	// 0, 1, 2
	Vetted int `json:"vetted"`
	// the number of seconds you wish to have the response cached
	CacheTTL int `json:"cache_ttl"`
}

func (q PageQuery) urlString() string {

	url := fmt.Sprintf(
		"http://eol.org/api/pages/1.0/%d.json?images=%d&sounds=%d&maps=%d&text=%d&iucn=%v",
		q.ID,
		q.Images,
		q.Sounds,
		q.Maps,
		q.Text,
		q.IUCN,
	)

	if q.Subjects != "" {
		url += fmt.Sprintf("&subjects=%s", q.Subjects)
	}
	if q.Licenses != "" {
		url += fmt.Sprintf("&licenses=%s", q.Licenses)
	}
	if q.CommonNames  {
		url += "&common_name=true"
	}
	if q.Details  {
		url += "&details=true"
	}
	if q.Synonyms  {
		url += "&synonyms=true"
	}
	if q.References {
		url += "&references=true"
	}

	if q.Vetted > 0 {
		url += fmt.Sprintf("vetted=%d", q.Vetted)
	}

	if q.CacheTTL > 0 {
		url += fmt.Sprintf("cache_ttl=%d", q.CacheTTL)
	}

	return url
}

type DataObject struct {
	Agents []struct {
		FullName string `json:"full_name"`
		Homepage string `json:"homepage"`
		Role     string `json:"role"`
	} `json:"agents"`
	Created             string        `json:"created"`
	DataObjectVersionID int           `json:"dataObjectVersionID"`
	DataRating          float64       `json:"dataRating"`
	DataSubtype         string        `json:"dataSubtype"`
	DataType            DataType        `json:"dataType"`
	Description         string        `json:"description"`
	Identifier          string        `json:"identifier"`
	Language            string        `json:"language"`
	License             string        `json:"license"`
	MimeType            string        `json:"mimeType"`
	MediaUrl        string          `json:"mediaURL"`
	EOLMediaUrl        string          `json:"eolMediaURL"`
	Modified            string        `json:"modified"`
	References          []interface{} `json:"references"`
	RightsHolder        string        `json:"rightsHolder"`
	Source              string        `json:"source"`
	VettedStatus        string        `json:"vettedStatus"`
	Subject        string        `json:"subject"`
	Height  int `json:"height"`
	Width  int `json:"width"`
	CropX  string `json:"crop_x"`
	CropY  string `json:"crop_y"`
	CropWidth  string `json:"crop_width"`
}

type PageResponse struct {
	DataObjects []DataObject `json:"dataObjects"`
	Identifier     int      `json:"identifier"`
	References     []string `json:"references"`
	RichnessScore  float64  `json:"richness_score"`
	ScientificName string   `json:"scientificName"`
	Synonyms       []struct {
		Relationship string `json:"relationship"`
		Resource     string `json:"resource"`
		Synonym      string `json:"synonym"`
	} `json:"synonyms"`
	TaxonConcepts []struct {
		CanonicalForm   string `json:"canonicalForm"`
		Identifier      int    `json:"identifier"`
		NameAccordingTo string `json:"nameAccordingTo"`
		ScientificName  string `json:"scientificName"`
		SourceIdentfier string `json:"sourceIdentfier"`
		TaxonRank       string `json:"taxonRank"`
	} `json:"taxonConcepts"`
	VernacularNames []struct {
		EolPreferred   bool   `json:"eol_preferred"`
		Language       string `json:"language"`
		VernacularName string `json:"vernacularName"`
	} `json:"vernacularNames"`
}

type DataType string
const (
	DataTypeText = DataType("http://purl.org/dc/dcmitype/Text")
	DataTypeStillImage = DataType("http://purl.org/dc/dcmitype/StillImage")
)

type Media struct {
	Source string `json:"source" bson:"source"`
	Value string `json:"value" bson:"value"`
}

func (this *PageResponse) Texts() (response []Media) {
	for _, o := range this.DataObjects {
		if o.DataType == DataTypeText && o.Description != "" {
			response = append(response, Media{
				Value: o.Description,
				Source: o.Source,
			})
		}
	}
	return
}

func (this *PageResponse) Images() (response []Media) {
	for _, o := range this.DataObjects {
		if o.DataType == DataTypeStillImage && o.EOLMediaUrl != "" {
			response = append(response, Media{
				Value: o.EOLMediaUrl,
				Source: o.Source,
			})
		}
	}
	return
}


func Page(q PageQuery) (*PageResponse, error) {

	if q.ID == 0 {
		return nil, errors.New("a page id is required")
	}

	resp, err := http.Get(q.urlString())
	if err != nil {
		return nil, errors.Wrap(err, "could not get http response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(errors.New(resp.Status), "StatusCode: %d; URL: %s", resp.StatusCode, q.urlString())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read http response body")
	}

	var response PageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal http response")
	}

	return &response, nil

}
