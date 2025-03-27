package genius

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"song-library/internal/domain/models"
	"song-library/internal/lib/logger/sl"
)

type TrackInfoSaver interface {
	Save(ctx context.Context, info *models.TrackInfo) error
}

type Client struct {
	log            *slog.Logger
	client         *http.Client
	baseURL        string
	accessToken    string
	trackInfoSaver TrackInfoSaver
}

var (
	ErrTrackNotFound  = errors.New("track not found")
	ErrLyricsNotFound = errors.New("lyrics not found")
)

func New(log *slog.Logger,
	baseURL string,
	accessToken string,
	trackInfoSaver TrackInfoSaver,
) *Client {
	return &Client{
		log:            log,
		baseURL:        baseURL,
		accessToken:    accessToken,
		client:         &http.Client{},
		trackInfoSaver: trackInfoSaver,
	}
}

type SearchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				ID                    int    `json:"id"`
				Title                 string `json:"title"`
				ArtistNames           string `json:"artist_names"`
				ReleaseDate           string `json:"release_date"`
				ReleaseDateComponents *struct {
					Year  int `json:"year"`
					Month int `json:"month"`
					Day   int `json:"day"`
				} `json:"release_date_components"`
				URL            string `json:"url"`
				PrimaryArtists []struct {
					Name string `json:"name"`
				} `json:"primary_artists"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

func (c *Client) TrackInfo(ctx context.Context,
	artist string,
	title string,
) error {
	const op = "service.genius.TrackInfo"

	log := c.log.With(slog.String("op", op),
		slog.String("artist", artist),
		slog.String("title", title),
	)

	log.Info("Fetching track info")

	req, err := c.buildGeniusSearchRequest(artist, title)
	if err != nil {
		c.log.Error("Failed to build search request", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("Request URL", slog.String("url", req.URL.String()))

	res, err := c.fetchTrackData(req)
	if err != nil {
		c.log.Error("Failed to fetch track data", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	trackInfo, err := c.extractTrackInfo(res)
	if err != nil {
		c.log.Error("Failed to extract track info", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.trackInfoSaver.Save(ctx, trackInfo); err != nil {
		c.log.Error("Failed to save track info", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) buildGeniusSearchRequest(artist, title string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("q", fmt.Sprintf("%s %s", artist, title))
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	return req, nil
}

func (c *Client) fetchTrackData(req *http.Request) (*SearchResponse, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res SearchResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if len(res.Response.Hits) == 0 {
		return nil, ErrTrackNotFound
	}

	return &res, nil
}

func (c *Client) extractTrackInfo(response *SearchResponse) (*models.TrackInfo, error) {
	trackInfo := response.Response.Hits[0].Result

	releaseDate := trackInfo.ReleaseDate
	if releaseDate == "" && trackInfo.ReleaseDateComponents != nil {
		components := trackInfo.ReleaseDateComponents
		releaseDate = fmt.Sprintf("%d-%02d-%02d", components.Year, components.Month, components.Day)
	}

	lyrics, err := c.parseLyrics(trackInfo.URL)
	if err != nil {
		c.log.Error("Failed to parse lyrics", sl.Err(err))

		return nil, err
	}

	return &models.TrackInfo{
		Title:       trackInfo.Title,
		Artist:      trackInfo.ArtistNames,
		ReleaseDate: releaseDate,
		Lyrics:      lyrics,
	}, nil
}

func (c *Client) parseLyrics(songURL string) (string, error) {
	resp, err := http.Get(songURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	lyrics := ""
	doc.Find("div[data-lyrics-container=true], .Lyrics__Container").Each(func(i int, s *goquery.Selection) {
		lyrics += strings.TrimSpace(s.Text()) + "\n\n"
	})

	if lyrics == "" {
		return "", ErrLyricsNotFound
	}

	return lyrics, nil
}
