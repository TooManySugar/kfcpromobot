package kfcapi

import (
	"errors"
	"regexp"
	"fmt"
	"io"
	"log"
	"net/url"
	"net/http"
	"os"
	"strings"
	"time"
)

type Promo struct {
	Value struct {
		ProductID           int `json:"productId"`
		ExternalIdentifiers struct {
			RkeeperCode string `json:"rkeeperCode"`
			RkeeperID   string `json:"rkeeperId"`
			SiteID      string `json:"siteId"`
		} `json:"externalIdentifiers"`
		Type        string `json:"type"`
		Translation struct {
			En struct {
				Description string `json:"description"`
				Disclaimer  string `json:"disclaimer"`
				Title       string `json:"title"`
			} `json:"en"`
			Ru struct {
				Description string `json:"description"`
				Disclaimer  string `json:"disclaimer"`
				Title       string `json:"title"`
			} `json:"ru"`
		} `json:"translation"`
		Price struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"price"`
		Taxes struct {
			Key1000 struct {
				Title struct {
					En string `json:"en"`
					Ru string `json:"ru"`
				} `json:"title"`
				Value int    `json:"value"`
				Unit  string `json:"unit"`
			} `json:"key--1000"`
		} `json:"taxes"`
		Attributes struct {
		} `json:"attributes"`
		Categories struct {
			Coupons []int `json:"coupons"`
		} `json:"categories"`
		Media struct {
			Image string `json:"image"`
		} `json:"media"`
		Flags struct {
			SiteSales []string `json:"siteSales"`
		} `json:"flags"`
		ModifierGroups []struct {
			ModifierGroupID int    `json:"modifierGroupId"`
			GroupType       string `json:"groupType"`
			Title           struct {
				En string `json:"en"`
				Ru string `json:"ru"`
			} `json:"title"`
			UpLimit   int `json:"upLimit"`
			DownLimit int `json:"downLimit"`
			Modifiers []struct {
				ModifierID          int64  `json:"modifierId"`
				ProductID           int    `json:"productId"`
				Type                string `json:"type"`
				ExternalIdentifiers struct {
					RkeeperID string `json:"rkeeperId"`
				} `json:"externalIdentifiers"`
				Title struct {
					En string `json:"en"`
					Ru string `json:"ru"`
				} `json:"title"`
				Media struct {
					Image string `json:"image"`
				} `json:"media"`
				Price struct {
					Amount   int    `json:"amount"`
					Currency string `json:"currency"`
				} `json:"price"`
				UpLimit   int `json:"upLimit"`
				SortOrder int `json:"sortOrder"`
			} `json:"modifiers"`
		} `json:"modifierGroups"`
		Recommendations   []int    `json:"recommendations"`
		OldPrice          int      `json:"oldPrice"`
		AvailableChannels []string `json:"availableChannels"`
	} `json:"value"`
	Status    int       `json:"status"`
	Elapsed   string    `json:"elapsed"`
	CreatedAt time.Time `json:"createdAt"`
}

var ErrNotFound = errors.New("kfcapi: code not found or expired")
var ErrExceedTryCountLimit = errors.New("kfcapi: exceed try count limit")

type Client struct {
	logger *log.Logger
	ApiLink string
}

func NewClientFromEnv(logger *log.Logger) (*Client, error) {
	apiLink := os.Getenv("KFCAPILINK")
	if len(apiLink) < 8 {
		return nil, errors.New("kfcapi: environment variable KFCAPILINK not set")
	}

	if _, err := url.Parse(apiLink); err != nil {
		return nil, errors.New("kfcapi: invalid api link " + err.Error())
	}

	return &Client {
		logger: log.New(logger.Writer(), "[KFC's API] ", log.LstdFlags | log.Lmsgprefix),
		ApiLink: apiLink,
	},
	nil
}

func (c *Client) apiGetValueWithRetry(url string) ([]byte, error) {
	const MAX_TRY_COUNT = 3
	var try_count int
	var resp *http.Response
	var err error
	for try_count = 0; try_count < MAX_TRY_COUNT; try_count++ {

		resp, err = http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			break
		}

		c.logger.Printf("api returned %d retrying...\n", resp.StatusCode)

		time.Sleep(time.Millisecond * 100)
	}

	if try_count >= MAX_TRY_COUNT {
		return nil, ErrExceedTryCountLimit
	}

	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Println("read responce body error: " + err.Error())
		return nil, err
	}

	if strings.HasPrefix(string(data),  `{"value":`) == false {
		dumpName := fmt.Sprintf("%s_%d.json", url, time.Now().Unix())
		c.logger.Println("responce body not have 'value', key dumping body", dumpName)
		f, err := os.Create(dumpName)
		if err == nil {
			defer f.Close()
			f.Write(data)
		} else {
			c.logger.Println("ERROR: can't create dump of responce body: ", err.Error())
		}

		return nil, ErrNotFound
	}

	return data, nil
}

func (c *Client) GetRestMenuRaw(restID string) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(c.ApiLink)
	sb.WriteString("/menu/api/v1/menu.short/")
	sb.WriteString(restID)
	sb.WriteString("/website/finger_lickin_good")
	url := sb.String()

	data, err := c.apiGetValueWithRetry(url)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) GetRestPromoCodes(restID string) ([]string, error) {
	data, err := c.GetRestMenuRaw(restID)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`"title"\s*:\s*"(Coupon|Купон) (\S+)"\s*,`)

	allcodes := re.FindAllSubmatch(data, -1)

	result := []string{}

	for _, v := range allcodes {
		result = append(result, (string)(v[2]))
	}

	return result, nil
}

func (c *Client) GetRestPromoInfoRaw(restID string, code string) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(c.ApiLink)
	sb.WriteString("/menu/api/v1/menu.product/")
	sb.WriteString(restID)
	sb.WriteString("/website/finger_lickin_good/")
	sb.WriteString(code)
	url := sb.String()

	data, err := c.apiGetValueWithRetry(url)
	if err != nil {
		return nil, err
	}

	return data, nil
}
