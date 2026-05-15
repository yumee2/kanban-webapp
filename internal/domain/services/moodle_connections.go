package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"student-kanban/internal/adapters/moodle"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type MoodleConnectionRepository interface {
	Upsert(connection *entity.MoodleConnection) error
	GetByUserID(userID uuid.UUID) (*entity.MoodleConnection, error)
}

type MoodleAuthClient interface {
	ExchangeCredentialsForToken(baseURL, username, password, service string) (string, error)
	GetSiteInfo(baseURL, token string) (*moodle.SiteInfo, error)
	GetUserCourses(baseURL, token string, moodleUserID int64) ([]moodle.Course, error)
	GetCourseContents(baseURL, token string, courseID int64) ([]moodle.CourseSection, error)
}

type MoodleBoardRepository interface {
	CreateBoard(board *entity.Board) (string, error)
}

type MoodleListRepository interface {
	FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]entity.List, error)
}

type MoodleCardRepository interface {
	Create(ctx context.Context, card *entity.Card) error
	GetMaxPosition(ctx context.Context, listID uuid.UUID) (float64, error)
}

type ConnectMoodleInput struct {
	BaseURL  string
	Username string
	Password string
	Service  string
}

type MoodleConnectionInfo struct {
	BaseURL          string `json:"base_url"`
	MoodleUserID     int64  `json:"moodle_user_id"`
	MoodleUsername   string `json:"moodle_username"`
	MoodleFullName   string `json:"moodle_full_name"`
	MoodleSiteName   string `json:"moodle_site_name"`
	ServiceShortName string `json:"service_short_name"`
}

type moodleConnectionService struct {
	repo      MoodleConnectionRepository
	client    MoodleAuthClient
	boardRepo MoodleBoardRepository
	listRepo  MoodleListRepository
	cardRepo  MoodleCardRepository
}

func NewMoodleConnectionService(
	repo MoodleConnectionRepository,
	client MoodleAuthClient,
	boardRepo MoodleBoardRepository,
	listRepo MoodleListRepository,
	cardRepo MoodleCardRepository,
) *moodleConnectionService {
	return &moodleConnectionService{
		repo:      repo,
		client:    client,
		boardRepo: boardRepo,
		listRepo:  listRepo,
		cardRepo:  cardRepo,
	}
}

func (s *moodleConnectionService) Connect(userID uuid.UUID, input ConnectMoodleInput) (*MoodleConnectionInfo, error) {
	baseURL, err := moodle.NormalizeBaseURL(input.BaseURL)
	if err != nil {
		return nil, err
	}

	service := strings.TrimSpace(input.Service)
	if service == "" {
		service = "moodle_mobile_app"
	}

	token, err := s.client.ExchangeCredentialsForToken(baseURL, input.Username, input.Password, service)
	if err != nil {
		return nil, err
	}

	siteInfo, err := s.client.GetSiteInfo(baseURL, token)
	if err != nil {
		return nil, err
	}

	encryptedToken, err := encryptSecret(token)
	if err != nil {
		return nil, err
	}

	connection := &entity.MoodleConnection{
		UserID:           userID,
		BaseURL:          baseURL,
		MoodleUserID:     siteInfo.UserID,
		MoodleUsername:   siteInfo.Username,
		MoodleFullName:   siteInfo.FullName,
		MoodleSiteName:   siteInfo.SiteName,
		EncryptedToken:   encryptedToken,
		TokenFingerprint: fingerprintToken(token),
		ServiceShortName: service,
	}

	if err := s.repo.Upsert(connection); err != nil {
		return nil, err
	}

	return &MoodleConnectionInfo{
		BaseURL:          connection.BaseURL,
		MoodleUserID:     connection.MoodleUserID,
		MoodleUsername:   connection.MoodleUsername,
		MoodleFullName:   connection.MoodleFullName,
		MoodleSiteName:   connection.MoodleSiteName,
		ServiceShortName: connection.ServiceShortName,
	}, nil
}

func (s *moodleConnectionService) GetConnection(userID uuid.UUID) (*MoodleConnectionInfo, error) {
	connection, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &MoodleConnectionInfo{
		BaseURL:          connection.BaseURL,
		MoodleUserID:     connection.MoodleUserID,
		MoodleUsername:   connection.MoodleUsername,
		MoodleFullName:   connection.MoodleFullName,
		MoodleSiteName:   connection.MoodleSiteName,
		ServiceShortName: connection.ServiceShortName,
	}, nil
}

func (s *moodleConnectionService) GetCourses(userID uuid.UUID) ([]moodle.Course, error) {
	connection, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	token, err := decryptSecret(connection.EncryptedToken)
	if err != nil {
		return nil, err
	}

	return s.client.GetUserCourses(connection.BaseURL, token, connection.MoodleUserID)
}

func (s *moodleConnectionService) ImportCourseAsBoard(ctx context.Context, userID uuid.UUID, courseID int64) (string, error) {
	connection, err := s.repo.GetByUserID(userID)
	if err != nil {
		return "", err
	}

	token, err := decryptSecret(connection.EncryptedToken)
	if err != nil {
		return "", err
	}

	courses, err := s.client.GetUserCourses(connection.BaseURL, token, connection.MoodleUserID)
	if err != nil {
		return "", err
	}

	var selectedCourse *moodle.Course
	for i := range courses {
		if courses[i].ID == courseID {
			selectedCourse = &courses[i]
			break
		}
	}
	if selectedCourse == nil {
		return "", entity.ErrMoodleCourseNotFound
	}

	sections, err := s.client.GetCourseContents(connection.BaseURL, token, courseID)
	if err != nil {
		return "", err
	}

	description := buildBoardDescription(selectedCourse)
	board := &entity.Board{
		OwnerID:     userID,
		Title:       courseDisplayName(*selectedCourse),
		Description: &description,
	}

	boardID, err := s.boardRepo.CreateBoard(board)
	if err != nil {
		return "", err
	}

	lists, err := s.listRepo.FindByBoardID(ctx, board.ID)
	if err != nil {
		return "", err
	}
	if len(lists) == 0 {
		return "", errors.New("доска создана без стандартных списков")
	}

	targetListID := lists[0].ID
	position, err := s.cardRepo.GetMaxPosition(ctx, targetListID)
	if err != nil {
		return "", err
	}

	for _, module := range flattenImportableModules(sections) {
		position++
		description := buildCardDescription(module)
		card := &entity.Card{
			ListID:      targetListID,
			Title:       truncateString(strings.TrimSpace(module.Name), 255),
			Description: description,
			DueDate:     moduleDueDate(module),
			Priority:    entity.PriorityMedium,
			Position:    position,
		}
		if err := s.cardRepo.Create(ctx, card); err != nil {
			return "", err
		}
	}

	return boardID, nil
}

func encryptSecret(value string) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptSecret(encoded string) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("invalid encrypted token")
	}

	nonce, payload := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func newGCM() (cipher.AEAD, error) {
	key := os.Getenv("MOODLE_TOKEN_ENCRYPTION_KEY")
	if key == "" {
		return nil, entity.ErrMoodleTokenKeyMissing
	}

	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return gcm, nil
}

func fingerprintToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)
var supportedImportModules = []string{"assign", "quiz"}

func flattenImportableModules(sections []moodle.CourseSection) []moodle.CourseModule {
	modules := make([]moodle.CourseModule, 0)
	for _, section := range sections {
		for _, module := range section.Modules {
			if !module.UserVisible {
				continue
			}
			if !slices.Contains(supportedImportModules, module.ModName) {
				continue
			}
			if strings.TrimSpace(module.Name) == "" {
				continue
			}
			modules = append(modules, module)
		}
	}
	return modules
}

func buildBoardDescription(course *moodle.Course) string {
	parts := make([]string, 0, 2)
	if summary := sanitizeText(course.Summary); summary != "" {
		parts = append(parts, summary)
	}
	if course.ViewURL != "" {
		parts = append(parts, "Moodle: "+course.ViewURL)
	}
	if len(parts) == 0 {
		return "Добавлено из Moodle"
	}
	return truncateString(strings.Join(parts, "\n\n"), 500)
}

func buildCardDescription(module moodle.CourseModule) *string {
	parts := []string{"Type: " + module.ModName}
	if description := sanitizeText(module.Description); description != "" {
		parts = append(parts, description)
	}
	if module.URL != "" {
		parts = append(parts, "Открыть в Moodle: "+module.URL)
	}

	description := truncateString(strings.Join(parts, "\n\n"), 500)
	return &description
}

func moduleDueDate(module moodle.CourseModule) *time.Time {
	for _, date := range module.Dates {
		label := strings.ToLower(strings.TrimSpace(date.Label))
		if strings.Contains(label, "due") || strings.Contains(label, "cut-off") || strings.Contains(label, "cutoff") {
			dueDate := time.Unix(date.Timestamp, 0)
			return &dueDate
		}
	}

	for _, date := range module.Dates {
		if date.Timestamp <= 0 {
			continue
		}
		dueDate := time.Unix(date.Timestamp, 0)
		return &dueDate
	}

	return nil
}

func courseDisplayName(course moodle.Course) string {
	if strings.TrimSpace(course.DisplayName) != "" {
		return truncateString(strings.TrimSpace(course.DisplayName), 100)
	}
	return truncateString(strings.TrimSpace(course.FullName), 100)
}

func sanitizeText(value string) string {
	cleaned := htmlTagPattern.ReplaceAllString(value, " ")
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	return strings.TrimSpace(cleaned)
}

func truncateString(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
