package site24x7

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSite24x7WebsiteMonitor() *schema.Resource {
	return &schema.Resource{
		Create: websiteMonitorCreate,
		Read:   websiteMonitorRead,
		Update: websiteMonitorUpdate,
		Delete: websiteMonitorDelete,
		Exists: websiteMonitorExists,

		Schema: map[string]*schema.Schema{
			"display_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"website": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"check_frequency": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},

			"http_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "G",
			},

			"auth_user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"auth_pass": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"matching_keyword_value": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"matching_keyword_severity": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},

			"unmatching_keyword_value": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"unmatching_keyword_severity": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},

			"match_regex_value": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"match_regex_severity": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},

			"match_case": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"user_agent": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"custom_headers": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},

			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},

			"location_profile_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"notification_profile_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"threshold_profile_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"monitor_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"user_group_ids": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			"action_ids": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			"action_alert_types": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			"use_name_server": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"third_party_services": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

type Status int

const (
	Down           Status = 0
	Up             Status = 1
	Trouble        Status = 2
	Suspended      Status = 5
	Maintenance    Status = 7
	Discovery      Status = 9
	DiscoveryError Status = 10
)

type ValueAndSeverity struct {
	Value    string `json:"value"`
	Severity Status `json:"severity"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ActionRef struct {
	ActionID  string `json:"action_id"`
	AlertType Status `json:"alert_type"`
}

type WebsiteMonitor struct {
	MonitorID             string           `json:"monitor_id,omitempty"`
	DisplayName           string           `json:"display_name"`
	Type                  string           `json:"type"`
	Website               string           `json:"website"`
	CheckFrequency        string           `json:"check_frequency"`
	HTTPMethod            string           `json:"http_method"`
	AuthUser              string           `json:"auth_user"`
	AuthPass              string           `json:"auth_pass"`
	MatchingKeyword       ValueAndSeverity `json:"matching_keyword"`
	UnmatchingKeyword     ValueAndSeverity `json:"unmatching_keyword"`
	MatchRegex            ValueAndSeverity `json:"match_regex"`
	MatchCase             bool             `json:"match_case"`
	UserAgent             string           `json:"user_agent"`
	CustomHeaders         []Header         `json:"custom_headers"`
	Timeout               int              `json:"timeout"`
	LocationProfileID     string           `json:"location_profile_id"`
	NotificationProfileID string           `json:"notification_profile_id"`
	ThresholdProfileID    string           `json:"threshold_profile_id"`
	MonitorGroupID        string           `json:"monitor_group_id,omitempty"`
	UserGroupIDs          []string         `json:"user_group_ids"`
	ActionIDs             []ActionRef      `json:"action_ids,omitempty"`
	UseNameServer         bool             `json:"use_name_server"`
	ThirdPartyServices    []string         `json:"third_party_services,omitempty"`
}

func websiteMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	return websiteMonitorCreateOrUpdate(http.MethodPost, "https://www.site24x7.com/api/monitors", http.StatusCreated, d, meta)
}

func websiteMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	return websiteMonitorCreateOrUpdate(http.MethodPut, "https://www.site24x7.com/api/monitors/"+d.Id(), http.StatusOK, d, meta)
}

func websiteMonitorCreateOrUpdate(method, url string, expectedResponseStatus int, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*http.Client)

	customHeaders := []Header{}
	for k, v := range d.Get("custom_headers").(map[string]interface{}) {
		customHeaders = append(customHeaders, Header{Name: k, Value: v.(string)})
	}

	var userGroupIDs []string
	for _, id := range d.Get("user_group_ids").([]interface{}) {
		userGroupIDs = append(userGroupIDs, id.(string))
	}

	actionIDs := d.Get("action_ids").([]interface{})
	actionAlertTypes := d.Get("action_alert_types").([]interface{})
	actionRefs := make([]ActionRef, len(actionIDs))
	for i := range actionRefs {
		alertType := Status(-1)
		if i < len(actionAlertTypes) {
			alertType = Status(actionAlertTypes[i].(int))
		}
		actionRefs[i] = ActionRef{ActionID: actionIDs[i].(string), AlertType: alertType}
	}

	var thirdPartyServiceIDs []string
	for _, id := range d.Get("third_party_services").([]interface{}) {
		thirdPartyServiceIDs = append(thirdPartyServiceIDs, id.(string))
	}

	m := &WebsiteMonitor{
		DisplayName:    d.Get("display_name").(string),
		Type:           "URL",
		Website:        d.Get("website").(string),
		CheckFrequency: strconv.Itoa(d.Get("check_frequency").(int)),
		HTTPMethod:     d.Get("http_method").(string),
		AuthUser:       d.Get("auth_user").(string),
		AuthPass:       d.Get("auth_pass").(string),
		MatchingKeyword: ValueAndSeverity{
			Value:    fixEmpty(d.Get("matching_keyword_value").(string)),
			Severity: Status(d.Get("matching_keyword_severity").(int)),
		},
		UnmatchingKeyword: ValueAndSeverity{
			Value:    fixEmpty(d.Get("unmatching_keyword_value").(string)),
			Severity: Status(d.Get("unmatching_keyword_severity").(int)),
		},
		MatchRegex: ValueAndSeverity{
			Value:    fixEmpty(d.Get("match_regex_value").(string)),
			Severity: Status(d.Get("match_regex_severity").(int)),
		},
		MatchCase:             d.Get("match_case").(bool),
		UserAgent:             d.Get("user_agent").(string),
		CustomHeaders:         customHeaders,
		Timeout:               d.Get("timeout").(int),
		LocationProfileID:     d.Get("location_profile_id").(string),
		NotificationProfileID: d.Get("notification_profile_id").(string),
		ThresholdProfileID:    d.Get("threshold_profile_id").(string),
		MonitorGroupID:        d.Get("monitor_group_id").(string),
		UserGroupIDs:          userGroupIDs,
		ActionIDs:             actionRefs,
		UseNameServer:         d.Get("use_name_server").(bool),
		ThirdPartyServices:    thirdPartyServiceIDs,
	}

	if m.LocationProfileID == "" {
		id, err := defaultLocationProfile(client)
		if err != nil {
			return err
		}
		m.LocationProfileID = id
		d.Set("location_profile_id", id)
	}
	if m.NotificationProfileID == "" {
		id, err := defaultNotificationProfile(client)
		if err != nil {
			return err
		}
		m.NotificationProfileID = id
		d.Set("notification_profile_id", id)
	}
	if m.ThresholdProfileID == "" {
		id, err := defaultThresholdProfile(client, "URL")
		if err != nil {
			return err
		}
		m.ThresholdProfileID = id
		d.Set("threshold_profile_id", id)
	}
	if len(m.UserGroupIDs) == 0 {
		id, err := defaultUserGroup(client)
		if err != nil {
			return err
		}
		m.UserGroupIDs = []string{id}
		d.Set("user_group_ids", []string{id})
	}

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedResponseStatus {
		return parseAPIError(resp.Body)
	}

	var apiResp struct {
		Data struct {
			MonitorID string `json:"monitor_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	d.SetId(apiResp.Data.MonitorID)
	// can't update the rest of the data here, because the response format is broken

	return nil
}

func fixEmpty(s string) string {
	if s == "" {
		return " "
	}
	return s
}

func websiteMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*http.Client)

	var apiResp struct {
		Data WebsiteMonitor `json:"data"`
	}
	if err := doGetRequest(client, "https://www.site24x7.com/api/monitors/"+d.Id(), &apiResp); err != nil {
		return err
	}
	updateWebsiteMonitorResourceData(d, &apiResp.Data)

	return nil
}

func updateWebsiteMonitorResourceData(d *schema.ResourceData, m *WebsiteMonitor) {
	d.Set("display_name", m.DisplayName)
	d.Set("website", m.Website)
	d.Set("check_frequency", m.CheckFrequency)
	d.Set("timeout", m.Timeout)
	d.Set("http_method", m.HTTPMethod)
	d.Set("auth_user", m.AuthUser)
	d.Set("auth_pass", m.AuthPass)
	d.Set("matching_keyword_value", m.MatchingKeyword.Value)
	d.Set("matching_keyword_severity", int(m.MatchingKeyword.Severity))
	d.Set("unmatching_keyword_value", m.UnmatchingKeyword.Value)
	d.Set("unmatching_keyword_severity", int(m.UnmatchingKeyword.Severity))
	d.Set("match_regex_value", m.MatchRegex.Value)
	d.Set("match_regex_severity", int(m.MatchRegex.Severity))
	d.Set("match_case", m.MatchCase)
	d.Set("user_agent", m.UserAgent)
	customHeaders := make(map[string]interface{})
	for _, h := range m.CustomHeaders {
		if h.Name == "" {
			continue
		}
		customHeaders[h.Name] = h.Value
	}
	d.Set("custom_headers", customHeaders)
	d.Set("location_profile_id", m.LocationProfileID)
	d.Set("notification_profile_id", m.NotificationProfileID)
	d.Set("threshold_profile_id", m.ThresholdProfileID)
	d.Set("monitor_group_id", m.MonitorGroupID)
	d.Set("user_group_ids", m.UserGroupIDs)
	actionIDs := make([]string, len(m.ActionIDs))
	actionAlertTypes := make([]Status, len(m.ActionIDs))
	for i, r := range m.ActionIDs {
		actionIDs[i] = r.ActionID
		actionAlertTypes[i] = r.AlertType
	}
	d.Set("action_ids", actionIDs)
	d.Set("action_alert_types", actionAlertTypes)
	d.Set("use_name_server", m.UseNameServer)
	d.Set("third_party_services", m.ThirdPartyServices)
}

func websiteMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*http.Client)

	req, err := http.NewRequest(http.MethodDelete, "https://www.site24x7.com/api/monitors/"+d.Id(), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseAPIError(resp.Body)
	}

	return nil
}

func websiteMonitorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return fetchWebsiteMonitorExists(meta.(*http.Client), d.Id())
}

func fetchWebsiteMonitorExists(client *http.Client, id string) (bool, error) {
	resp, err := client.Get("https://www.site24x7.com/api/monitors/" + id)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, parseAPIError(resp.Body)
	}
}

func defaultLocationProfile(client *http.Client) (string, error) {
	var apiResp struct {
		Data []struct {
			ProfileID string `json:"profile_id"`
		} `json:"data"`
	}
	if err := doGetRequest(client, "https://www.site24x7.com/api/location_profiles", &apiResp); err != nil {
		return "", err
	}
	return apiResp.Data[0].ProfileID, nil
}

func defaultNotificationProfile(client *http.Client) (string, error) {
	var apiResp struct {
		Data []struct {
			ProfileID string `json:"profile_id"`
		} `json:"data"`
	}
	if err := doGetRequest(client, "https://www.site24x7.com/api/notification_profiles", &apiResp); err != nil {
		return "", err
	}
	return apiResp.Data[0].ProfileID, nil
}

func defaultThresholdProfile(client *http.Client, monitorType string) (string, error) {
	var apiResp struct {
		Data []struct {
			ProfileID   string `json:"profile_id"`
			MonitorType string `json:"type"`
		} `json:"data"`
	}
	if err := doGetRequest(client, "https://www.site24x7.com/api/threshold_profiles", &apiResp); err != nil {
		return "", err
	}
	for _, p := range apiResp.Data {
		if p.MonitorType == monitorType {
			return p.ProfileID, nil
		}
	}
	return "", errors.New("no threshold profile found")
}

func defaultUserGroup(client *http.Client) (string, error) {
	var apiResp struct {
		Data []struct {
			UserGroupID string `json:"user_group_id"`
		} `json:"data"`
	}
	if err := doGetRequest(client, "https://www.site24x7.com/api/user_groups", &apiResp); err != nil {
		return "", err
	}
	return apiResp.Data[0].UserGroupID, nil
}

func doGetRequest(client *http.Client, url string, data interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseAPIError(resp.Body)
	}

	return json.NewDecoder(resp.Body).Decode(data)
}

func parseAPIError(r io.Reader) error {
	var apiErr struct {
		ErrorCode int             `json:"error_code"`
		Message   string          `json:"message"`
		ErrorInfo json.RawMessage `json:"error_info"`
	}
	if err := json.NewDecoder(r).Decode(&apiErr); err != nil {
		return err
	}
	if len(apiErr.ErrorInfo) != 0 {
		return fmt.Errorf("%s (%s)", apiErr.Message, string(apiErr.ErrorInfo))
	}
	return fmt.Errorf("%s", apiErr.Message)
}
