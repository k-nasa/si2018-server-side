package user

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/eure/si2018-server-side/entities"
	"github.com/eure/si2018-server-side/repositories"
	si "github.com/eure/si2018-server-side/restapi/summerintern"
)

func GetUsers(p si.GetUsersParams) middleware.Responder {
	//FIXME Token認証は共通化したい
	r := repositories.NewUserTokenRepository()

	ent, err := r.GetByToken(p.Token)

	if err != nil {
		return getUsersInteralServerErrorResponse("Internal Server Error")
	}
	if ent == nil {
		return getUsersUnauthorizedResponse("Your Token Is Invalid")
	}

	userToken := ent.Build()
	userID := userToken.UserID

	userLikeRepository := repositories.NewUserLikeRepository()
	exclusionIds, err := userLikeRepository.FindLikeAll(userID)

	if err != nil {
		return getUsersInteralServerErrorResponse("Internal Server Error")
	}

	userRepository := repositories.NewUserRepository()
	userEnt, err := userRepository.GetByUserID(userID)
	if err != nil {
		return getUsersInteralServerErrorResponse("Internal Server Error")
	}

	// int64になっているのでcastする必要がある
	limit := int(p.Limit)
	offset := int(p.Offset)
	gender := userEnt.GetOppositeGender()

	var usersEnt entities.Users
	usersEnt, err = userRepository.FindUsers(limit, offset, gender, exclusionIds)
	if err != nil {
		return getUsersInteralServerErrorResponse("Internal Server Error")
	}

	var ids []int64
	for _, u := range usersEnt {
		ids = append(ids, u.ID)
	}

	userImageRepository := repositories.NewUserImageRepository()
	images, err := userImageRepository.GetByUserIDs(ids)
	if err != nil {
		return getUsersInteralServerErrorResponse("Internal Server Error")
	}

	for i := range usersEnt {
		usersEnt[i].ImageURI = images[i].Path
	}

	users := usersEnt.Build()
	return si.NewGetUsersOK().WithPayload(users)
}

func GetProfileByUserID(p si.GetProfileByUserIDParams) middleware.Responder {
	//FIXME Token認証は共通化したい
	r := repositories.NewUserTokenRepository()

	userTokenEnt, err := r.GetByToken(p.Token)

	if err != nil {
		return getUserProfileByUserIDInternalServerErrorResponse("Internal Server Error")
	}
	if userTokenEnt == nil {
		return getUsersUnauthorizedResponse("Your Token Is Invalid")
	}

	userID := p.UserID

	userEnt, err := repositories.NewUserRepository().GetByUserID(userID)
	if err != nil {
		return getUserProfileByUserIDInternalServerErrorResponse("Internal Server Error")
	}

	if userEnt == nil {
		return getUserProfileByUserIDNotFoundResponse("User Not Found")
	}

	userImageRepository := repositories.NewUserImageRepository()
	userImage, err := userImageRepository.GetByUserID(userID)
	if err != nil {
		return getUserProfileByUserIDInternalServerErrorResponse("Internal Server Error")
	}

	userEnt.ImageURI = userImage.Path

	user := userEnt.Build()
	return si.NewGetProfileByUserIDOK().WithPayload(&user)
}

func PutProfile(p si.PutProfileParams) middleware.Responder {
	userTokenRepository := repositories.NewUserTokenRepository()
	userTokenEnt, err := userTokenRepository.GetByToken(p.Params.Token)

	if err != nil {
		return putProfileInternalServerErrorResponse()
	}

	if userTokenEnt == nil {
		return putProfileUnauthorizedResponse()
	}

	userID := p.UserID

	if userTokenEnt.UserID != userID {
		return putProfileForbiddenResponse()
	}

	userEnt, err := repositories.NewUserRepository().GetByUserID(userID)

	if err != nil || userEnt == nil {
		return putProfileInternalServerErrorResponse()
	}

	updateUserEnt := repositories.NewUserRepository().ParamsToUserEnt(userEnt, p.Params)

	err = repositories.NewUserRepository().Update(updateUserEnt)
	if err != nil {
		return putProfileInternalServerErrorResponse()
	}

	updatedUserEnt, err := repositories.NewUserRepository().GetByUserID(userID)
	if err != nil {
		return putProfileInternalServerErrorResponse()
	}

	user := updatedUserEnt.Build()
	return si.NewPutProfileOK().WithPayload(&user)
}

// ここからResponderを返す関数群
func getUsersInteralServerErrorResponse(message string) middleware.Responder {
	return si.NewGetUsersInternalServerError().WithPayload(
		&si.GetUsersInternalServerErrorBody{
			Code:    "500",
			Message: message,
		})
}

func getUsersUnauthorizedResponse(message string) middleware.Responder {
	return si.NewGetUsersUnauthorized().WithPayload(
		&si.GetUsersUnauthorizedBody{
			Code:    "401",
			Message: message,
		})
}

func getUserProfileByUserIDInternalServerErrorResponse(message string) middleware.Responder {
	return si.NewGetProfileByUserIDInternalServerError().WithPayload(
		&si.GetProfileByUserIDInternalServerErrorBody{
			Code:    "500",
			Message: message,
		})
}

func getUserProfileByUserIDNotFoundResponse(message string) middleware.Responder {
	return si.NewGetProfileByUserIDNotFound().WithPayload(
		&si.GetProfileByUserIDNotFoundBody{
			Code:    "404",
			Message: message,
		})
}

func getUserProfileByUserIDUnauthorizeResponse(message string) middleware.Responder {
	return si.NewGetProfileByUserIDUnauthorized().WithPayload(
		&si.GetProfileByUserIDUnauthorizedBody{
			Code:    "401",
			Message: message,
		})
}

func putProfileInternalServerErrorResponse() middleware.Responder {
	return si.NewPutProfileInternalServerError().WithPayload(
		&si.PutProfileInternalServerErrorBody{
			Code:    "500",
			Message: "Internal Server Error",
		})
}

func putProfileUnauthorizedResponse() middleware.Responder {
	return si.NewPutProfileUnauthorized().WithPayload(
		&si.PutProfileUnauthorizedBody{
			Code:    "401",
			Message: "Your Token Is Invalid",
		})
}

func putProfileForbiddenResponse() middleware.Responder {
	return si.NewPutProfileForbidden().WithPayload(
		&si.PutProfileForbiddenBody{
			Code:    "403",
			Message: "Forbidden",
		})
}
