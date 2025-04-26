package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// OAuth2用户信息
type UserInfo struct {
	ID        string
	Name      string
	Email     string
	AvatarURL string
}

// OAuth2提供商接口
type Provider interface {
	// 获取授权URL
	GetAuthURL(redirectURI, state string) string
	// 通过授权码获取Token
	ExchangeToken(ctx context.Context, code, redirectURI string) (accessToken string, err error)
	// 获取用户信息
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
	// 提供商名称
	Name() string
}

// GitHub OAuth2提供商
type GithubProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// NewGithubProvider 创建GitHub提供商
func NewGithubProvider(clientID, clientSecret string) *GithubProvider {
	return &GithubProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GetAuthURL 获取GitHub授权URL
func (p *GithubProvider) GetAuthURL(redirectURI, state string) string {
	u := "https://github.com/login/oauth/authorize"
	q := url.Values{}
	q.Add("client_id", p.ClientID)
	q.Add("redirect_uri", redirectURI)
	q.Add("scope", "user:email")
	q.Add("state", state)

	return fmt.Sprintf("%s?%s", u, q.Encode())
}

// ExchangeToken 通过授权码获取GitHub Token
func (p *GithubProvider) ExchangeToken(ctx context.Context, code, redirectURI string) (string, error) {
	u := "https://github.com/login/oauth/access_token"

	data := url.Values{}
	data.Add("client_id", p.ClientID)
	data.Add("client_secret", p.ClientSecret)
	data.Add("code", code)
	data.Add("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取GitHub Token失败, 状态码: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("GitHub Token为空")
	}

	return tokenResp.AccessToken, nil
}

// GetUserInfo 获取GitHub用户信息
func (p *GithubProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		hlog.Errorf("获取GitHub用户信息失败: %s", string(body))
		return nil, fmt.Errorf("获取GitHub用户信息失败, 状态码: %d", resp.StatusCode)
	}

	var userResp struct {
		ID     int    `json:"id"`
		Login  string `json:"login"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}

	// 如果Email为空，尝试获取用户邮箱
	var email string
	if userResp.Email == "" {
		email = p.getGithubEmail(ctx, accessToken)
	} else {
		email = userResp.Email
	}

	userInfo := &UserInfo{
		ID:        fmt.Sprintf("%d", userResp.ID),
		Name:      userResp.Name,
		Email:     email,
		AvatarURL: userResp.Avatar,
	}

	return userInfo, nil
}

// getGithubEmail 获取GitHub用户邮箱
func (p *GithubProvider) getGithubEmail(ctx context.Context, accessToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		hlog.Errorf("创建获取GitHub邮箱请求失败: %v", err)
		return ""
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		hlog.Errorf("获取GitHub邮箱失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hlog.Errorf("获取GitHub邮箱失败, 状态码: %d", resp.StatusCode)
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		hlog.Errorf("解析GitHub邮箱失败: %v", err)
		return ""
	}

	// 优先使用主邮箱
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}

	// 如果没有主邮箱，使用第一个验证过的邮箱
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}

	// 如果没有验证过的邮箱，使用第一个邮箱
	if len(emails) > 0 {
		return emails[0].Email
	}

	return ""
}

// Name 获取提供商名称
func (p *GithubProvider) Name() string {
	return "github"
}

// GoogleProvider OAuth2提供商
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
}

// NewGoogleProvider 创建Google提供商
func NewGoogleProvider(clientID, clientSecret string) *GoogleProvider {
	return &GoogleProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GetAuthURL 获取Google授权URL
func (p *GoogleProvider) GetAuthURL(redirectURI, state string) string {
	u := "https://accounts.google.com/o/oauth2/v2/auth"
	q := url.Values{}
	q.Add("client_id", p.ClientID)
	q.Add("redirect_uri", redirectURI)
	q.Add("response_type", "code")
	q.Add("scope", "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email")
	q.Add("state", state)

	return fmt.Sprintf("%s?%s", u, q.Encode())
}

// ExchangeToken 通过授权码获取Google Token
func (p *GoogleProvider) ExchangeToken(ctx context.Context, code, redirectURI string) (string, error) {
	u := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Add("client_id", p.ClientID)
	data.Add("client_secret", p.ClientSecret)
	data.Add("code", code)
	data.Add("redirect_uri", redirectURI)
	data.Add("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取Google Token失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("Google Token为空")
	}

	return tokenResp.AccessToken, nil
}

// GetUserInfo 获取Google用户信息
func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取Google用户信息失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var userResp struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}

	userInfo := &UserInfo{
		ID:        userResp.ID,
		Name:      userResp.Name,
		Email:     userResp.Email,
		AvatarURL: userResp.Picture,
	}

	return userInfo, nil
}

// Name 获取提供商名称
func (p *GoogleProvider) Name() string {
	return "google"
}

// GetProvider 获取OAuth2提供商
func GetProvider(provider, clientID, clientSecret string) (Provider, error) {
	switch provider {
	case "github":
		return NewGithubProvider(clientID, clientSecret), nil
	case "google":
		return NewGoogleProvider(clientID, clientSecret), nil
	default:
		return nil, fmt.Errorf("不支持的OAuth2提供商: %s", provider)
	}
}
