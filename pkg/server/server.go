package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/zhashkevych/go-pocket-sdk"
	"tg-giga/pkg/repository"
)

const (
	PORT = ":80"
)

type AuthorizationServer struct {
	server          *http.Server
	pocketClient    *pocket.Client
	tokenRepository repository.TokenRepository
	redirectURL     string
}

func NewAuthorizationServer(pocketClient *pocket.Client, tokenRepository repository.TokenRepository, redirectURL string) *AuthorizationServer {
	return &AuthorizationServer{pocketClient: pocketClient, tokenRepository: tokenRepository, redirectURL: redirectURL}
}

func (s *AuthorizationServer) Start() error {
	s.server = &http.Server{
		Addr:    PORT,
		Handler: s, // Handler - struct, который содержит метод ServeHTTP(ResponseWriter, *Request)
	}

	return s.server.ListenAndServe()
}

// ServeHTTP будет вызываться каждый раз, когда на сервер будет приходить запрос
// Сюда будут прилетать GET-запросы от редиректов, в котором будет указать ?chat_id=<my_chat_id>
// По этому ChatID в базе будет найден RequestToken
// Далее с помощью RequestToken идёт запрос на PocketAPI на получение AccessToken. Сохранить в БД этот токен
func (s *AuthorizationServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	chatIDParam := r.URL.Query().Get("chat_id")

	if chatIDParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Сам RequestTokens генерируется на этапе "generateAuthorizationLink"
	requestToken, err := s.tokenRepository.Get(chatID, repository.RequestTokens)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized) // На случай, если токен не был найден
		return
	}

	// r.Context() - Request Контекст
	authResp, err := s.pocketClient.Authorize(r.Context(), requestToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	saveAtErr := s.tokenRepository.Save(chatID, authResp.AccessToken, repository.AccessTokens)
	if saveAtErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("chat_id: %d\nrequest_token: %s\naccess_token: %s\n", chatID, requestToken, authResp.AccessToken)

	// INFO: Header "Location" и StatusCode 301 вынудит браузер сделать Redirect по переданному адресу
	w.Header().Add("Location", s.redirectURL)
	w.WriteHeader(http.StatusMovedPermanently)
}
