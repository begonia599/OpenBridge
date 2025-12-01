package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// User 用户信息
type User struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"` // SHA256 hash
	Email     string    `json:"email"`
	APIKeys   []UserKey `json:"api_keys"`
	CreatedAt time.Time `json:"created_at"`
	Disabled  bool      `json:"disabled"`
}

// UserKey 用户的 API Key
type UserKey struct {
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used,omitempty"`
	Usage     int64     `json:"usage"` // 使用次数
}

// UserStore 用户存储
type UserStore struct {
	Users    map[string]*User `json:"users"` // username -> User
	Sessions map[string]*Session `json:"-"` // session_id -> Session
	mu       sync.RWMutex
	filePath string
}

// Session 用户会话
type Session struct {
	SessionID string    `json:"session_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

var store *UserStore

// Init 初始化用户存储
func Init(filepath string) error {
	store = &UserStore{
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
		filePath: filepath,
	}

	// 尝试加载已有用户
	data, err := os.ReadFile(filepath)
	if err != nil {
		// 文件不存在，创建默认管理员
		if os.IsNotExist(err) {
			return store.createDefaultUser()
		}
		return err
	}

	if err := json.Unmarshal(data, &store.Users); err != nil {
		return err
	}

	return nil
}

// createDefaultUser 创建默认用户
func (s *UserStore) createDefaultUser() error {
	// 创建默认测试用户
	defaultUser := &User{
		Username:  "demo",
		Password:  hashPassword("demo123"),
		Email:     "demo@openbridge.local",
		APIKeys:   []UserKey{},
		CreatedAt: time.Now(),
		Disabled:  false,
	}

	// 为默认用户生成一个 API Key
	key := generateAPIKey()
	defaultUser.APIKeys = append(defaultUser.APIKeys, UserKey{
		Key:       key,
		Name:      "默认 Key",
		CreatedAt: time.Now(),
		Usage:     0,
	})

	s.mu.Lock()
	s.Users[defaultUser.Username] = defaultUser
	s.mu.Unlock()

	return s.save()
}

// Register 注册新用户
func (s *UserStore) Register(username, password, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Users[username]; exists {
		return fmt.Errorf("用户名已存在")
	}

	user := &User{
		Username:  username,
		Password:  hashPassword(password),
		Email:     email,
		APIKeys:   []UserKey{},
		CreatedAt: time.Now(),
		Disabled:  false,
	}

	// 自动为新用户生成一个 API Key
	key := generateAPIKey()
	user.APIKeys = append(user.APIKeys, UserKey{
		Key:       key,
		Name:      "默认 Key",
		CreatedAt: time.Now(),
		Usage:     0,
	})

	s.Users[username] = user
	return s.save()
}

// Login 用户登录
func (s *UserStore) Login(username, password string) (*Session, error) {
	s.mu.RLock()
	user, exists := s.Users[username]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	if user.Disabled {
		return nil, fmt.Errorf("账户已被禁用")
	}

	if user.Password != hashPassword(password) {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 创建会话
	session := &Session{
		SessionID: generateSessionID(),
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时有效期
	}

	s.mu.Lock()
	s.Sessions[session.SessionID] = session
	s.mu.Unlock()

	return session, nil
}

// Logout 用户登出
func (s *UserStore) Logout(sessionID string) {
	s.mu.Lock()
	delete(s.Sessions, sessionID)
	s.mu.Unlock()
}

// GetSession 获取会话
func (s *UserStore) GetSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	session, exists := s.Sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	if time.Now().After(session.ExpiresAt) {
		s.Logout(sessionID)
		return nil, fmt.Errorf("会话已过期")
	}

	return session, nil
}

// GetUser 获取用户信息
func (s *UserStore) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.Users[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	return user, nil
}

// GenerateAPIKey 为用户生成新的 API Key
func (s *UserStore) GenerateAPIKey(username, keyName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.Users[username]
	if !exists {
		return "", fmt.Errorf("用户不存在")
	}

	key := generateAPIKey()
	userKey := UserKey{
		Key:       key,
		Name:      keyName,
		CreatedAt: time.Now(),
		Usage:     0,
	}

	user.APIKeys = append(user.APIKeys, userKey)
	s.save()

	return key, nil
}

// DeleteAPIKey 删除用户的 API Key
func (s *UserStore) DeleteAPIKey(username, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.Users[username]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	for i, k := range user.APIKeys {
		if k.Key == key {
			user.APIKeys = append(user.APIKeys[:i], user.APIKeys[i+1:]...)
			s.save()
			return nil
		}
	}

	return fmt.Errorf("Key 不存在")
}

// ValidateAPIKey 验证 API Key 并返回用户名
func (s *UserStore) ValidateAPIKey(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for username, user := range s.Users {
		if user.Disabled {
			continue
		}
		for _, k := range user.APIKeys {
			if k.Key == key {
				return username, true
			}
		}
	}

	return "", false
}

// RecordKeyUsage 记录 API Key 使用
func (s *UserStore) RecordKeyUsage(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.Users {
		for i, k := range user.APIKeys {
			if k.Key == key {
				user.APIKeys[i].Usage++
				user.APIKeys[i].LastUsed = time.Now()
				// 异步保存，避免阻塞
				go s.save()
				return
			}
		}
	}
}

// save 保存用户数据
func (s *UserStore) save() error {
	data, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

// GetStore 获取用户存储实例
func GetStore() *UserStore {
	return store
}

// hashPassword 使用 SHA256 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// generateAPIKey 生成 API Key
func generateAPIKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "sk-user-" + hex.EncodeToString(bytes)
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// SetupRoutes 设置用户相关路由
func SetupRoutes(r *gin.Engine) {
	// 公开路由
	r.GET("/user", serveUserPage)
	r.POST("/user/api/register", handleRegister)
	r.POST("/user/api/login", handleLogin)
	r.POST("/user/api/logout", handleLogout)

	// 需要认证的路由
	user := r.Group("/user/api")
	user.Use(authMiddleware())
	{
		user.GET("/profile", getProfile)
		user.GET("/keys", listKeys)
		user.POST("/keys/generate", generateKey)
		user.DELETE("/keys/:key", deleteKey)
		user.GET("/usage", getUsage)
	}
}

// authMiddleware 用户认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		session, err := store.GetSession(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("username", session.Username)
		c.Next()
	}
}

// 处理函数
func handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名至少3个字符，密码至少6个字符"})
		return
	}

	if err := store.Register(req.Username, req.Password, req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	session, err := store.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 设置 cookie
	c.SetCookie("session_id", session.SessionID, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message":  "登录成功",
		"username": session.Username,
	})
}

func handleLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("session_id")
	if sessionID != "" {
		store.Logout(sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "已登出"})
}

func getProfile(c *gin.Context) {
	username := c.GetString("username")
	user, err := store.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"key_count":  len(user.APIKeys),
	})
}

func listKeys(c *gin.Context) {
	username := c.GetString("username")
	user, err := store.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": user.APIKeys})
}

func generateKey(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	c.ShouldBindJSON(&req)

	if req.Name == "" {
		req.Name = "API Key " + time.Now().Format("2006-01-02")
	}

	username := c.GetString("username")
	key, err := store.GenerateAPIKey(username, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":     key,
		"message": "Key 生成成功",
	})
}

func deleteKey(c *gin.Context) {
	username := c.GetString("username")
	key := c.Param("key")

	if err := store.DeleteAPIKey(username, key); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key 已删除"})
}

func getUsage(c *gin.Context) {
	username := c.GetString("username")
	user, err := store.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var totalUsage int64
	for _, key := range user.APIKeys {
		totalUsage += key.Usage
	}

	c.JSON(http.StatusOK, gin.H{
		"total_usage": totalUsage,
		"keys":        user.APIKeys,
	})
}

func serveUserPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(userHTML))
}

