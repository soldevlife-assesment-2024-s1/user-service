package usecases

import (
	"context"
	"fmt"

	"user-service/internal/module/user/models/entity"
	"user-service/internal/module/user/models/request"
	"user-service/internal/module/user/models/response"
	"user-service/internal/module/user/repositories"
	"user-service/internal/pkg/helpers"
	"user-service/internal/pkg/helpers/errors"
	"user-service/internal/pkg/log"

	"github.com/dgrijalva/jwt-go"
)

type usecases struct {
	repositories repositories.Repositories
	log          log.Logger
}

// GetProfile implements Usecases.
func (u *usecases) GetProfile(ctx context.Context, payload *request.GetProfile) (response.GetProfileResponse, error) {
	// check if record exists
	profile, err := u.repositories.FindProfileByUserID(ctx, payload.UserID)
	if err != nil {
		return response.GetProfileResponse{}, errors.InternalServerError(fmt.Sprintf("error finding profile by id: %s", err.Error()))
	}

	resp := response.GetProfileResponse{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Address:        profile.Address,
		District:       profile.District,
		City:           profile.City,
		State:          profile.State,
		Country:        profile.Country,
		Region:         profile.Region,
		Phone:          profile.Phone,
		PersonalID:     profile.PersonalID,
		TypePersonalID: profile.TypePersonalID,
	}

	return resp, nil
}

// GetUser implements Usecases.
func (u *usecases) GetUser(ctx context.Context, payload *request.GetUser) (response.GetUserResponse, error) {
	// check if record exists
	user, err := u.repositories.FindUserByID(ctx, payload.ID)
	if err != nil {
		return response.GetUserResponse{}, errors.InternalServerError(fmt.Sprintf("error finding user by id: %s", err.Error()))
	}

	resp := response.GetUserResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	return resp, nil
}

// Login implements Usecases.
func (u *usecases) Login(ctx context.Context, payload *request.Login) (response.LoginResponse, error) {
	// check if user exists
	user, err := u.repositories.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		return response.LoginResponse{}, errors.BadRequest(fmt.Sprintf("Invalid email or password %s", err.Error()))
	}

	if user.ID == 0 {
		u.log.Error(ctx, "user not found", nil)
		return response.LoginResponse{}, errors.BadRequest("Invalid email or password")
	}

	// check if password is correct
	if err := helpers.CheckPasswordHash(payload.Password, user.Password); err != nil {
		u.log.Error(ctx, "invalid password", err)
		return response.LoginResponse{}, errors.BadRequest("Invalid email or password")
	}

	// generate token
	token, refreshToken, expiredAt, err := helpers.GenerateToken(user.ID)
	if err != nil {
		return response.LoginResponse{}, errors.InternalServerError(fmt.Sprintf("error generating token: %s", err.Error()))
	}

	resp := response.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiredAt:    expiredAt.Unix(),
	}

	return resp, nil
}

// Register implements Usecases.
func (u *usecases) Register(ctx context.Context, payload *request.Register) error {
	// check if user already exists
	userExisting, err := u.repositories.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("error finding user by email: %s", err.Error()))
	}

	if userExisting.ID != 0 {
		return errors.BadRequest("user already exists")
	}

	// hash password
	hashedPassword, err := helpers.HashPassword(payload.Password)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("error hashing password: %s", err.Error()))
	}

	// create user
	user := entity.User{
		Email:    payload.Email,
		Password: hashedPassword,
	}

	if err := u.repositories.UpsertUser(ctx, &user); err != nil {
		return errors.InternalServerError(fmt.Sprintf("error upserting user: %s", err.Error()))
	}

	return nil
}

// CreateProfile implements Usecases.
func (u *usecases) CreateProfile(ctx context.Context, payload *request.CreateProfile) error {
	// check if user exists
	userExisting, err := u.repositories.FindUserByID(ctx, payload.UserID)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("error finding user by id: %s", err.Error()))
	}

	if userExisting.ID == 0 {
		return errors.BadRequest("user not found")
	}

	// create profile
	profile := entity.Profile{
		UserID:         payload.UserID,
		Address:        payload.Address,
		District:       payload.District,
		City:           payload.City,
		State:          payload.State,
		Country:        payload.Country,
		Region:         payload.Region,
		Phone:          payload.Phone,
		PersonalID:     payload.PersonalID,
		TypePersonalID: payload.TypePersonalID,
	}

	if err := u.repositories.UpsertProfile(ctx, &profile); err != nil {
		return errors.InternalServerError(fmt.Sprintf("error upserting profile: %s", err.Error()))
	}

	return nil
}

// UpdateProfile implements Usecases.
func (u *usecases) UpdateProfile(ctx context.Context, payload *request.UpdateProfile) error {
	// check if record exists
	profileExisting, err := u.repositories.FindProfileByUserID(ctx, payload.UserID)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("error finding profile by id: %s", err.Error()))
	}

	// update profile
	profile := entity.Profile{
		ID:             payload.ID,
		UserID:         profileExisting.UserID,
		Address:        payload.Address,
		District:       payload.District,
		City:           payload.City,
		State:          payload.State,
		Country:        payload.Country,
		Region:         payload.Region,
		Phone:          payload.Phone,
		PersonalID:     payload.PersonalID,
		TypePersonalID: payload.TypePersonalID,
	}

	if err := u.repositories.UpsertProfile(ctx, &profile); err != nil {
		return errors.InternalServerError(fmt.Sprintf("error upserting profile: %s", err.Error()))
	}

	return nil
}

// UpdateUser implements Usecases.
func (u *usecases) UpdateUser(ctx context.Context, payload *request.UpdateUser) error {
	// check if record exists
	userExisting, err := u.repositories.FindUserByID(ctx, payload.ID)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("error finding user by id: %s", err.Error()))
	}

	// update user
	user := entity.User{
		ID:       payload.ID,
		Email:    payload.Email,
		Password: userExisting.Password,
	}

	if err := u.repositories.UpsertUser(ctx, &user); err != nil {
		return errors.InternalServerError(fmt.Sprintf("error upserting user: %s", err.Error()))
	}

	return nil
}

// ValidateToken implements Usecases.
func (u *usecases) ValidateToken(ctx context.Context, payload *request.ValidateToken) (response.GetUserResponse, error) {
	tokenString := payload.Token
	// Define the secret key
	var secret = "your-secret-key"
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return response.GetUserResponse{}, errors.UnauthorizedError("invalid token")
	}

	// Extract the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return response.GetUserResponse{}, errors.UnauthorizedError("invalid token")
	}

	// Extract the user ID
	userID, ok := claims["userID"].(int)
	if !ok {
		return response.GetUserResponse{}, errors.UnauthorizedError("invalid token")
	}

	// check if user exists
	user, err := u.repositories.FindUserByID(ctx, userID)
	if err != nil {
		return response.GetUserResponse{}, errors.InternalServerError(fmt.Sprintf("error finding user by id: %s", err.Error()))
	}

	if user.ID == 0 {
		return response.GetUserResponse{}, errors.UnauthorizedError("invalid token")
	}

	response := response.GetUserResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	// Return the user ID
	return response, nil
}

type Usecases interface {
	Register(ctx context.Context, payload *request.Register) error
	Login(ctx context.Context, payload *request.Login) (response.LoginResponse, error)
	GetUser(ctx context.Context, payload *request.GetUser) (response.GetUserResponse, error)
	UpdateUser(ctx context.Context, payload *request.UpdateUser) error
	ValidateToken(ctx context.Context, payload *request.ValidateToken) (response.GetUserResponse, error)
	CreateProfile(ctx context.Context, payload *request.CreateProfile) error
	GetProfile(ctx context.Context, payload *request.GetProfile) (response.GetProfileResponse, error)
	UpdateProfile(ctx context.Context, payload *request.UpdateProfile) error
}

func New(repositories repositories.Repositories, log log.Logger) Usecases {
	return &usecases{
		repositories: repositories,
		log:          log,
	}
}
