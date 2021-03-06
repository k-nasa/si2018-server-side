package usermatch

import (
	"github.com/eure/si2018-server-side/entities"
	"github.com/eure/si2018-server-side/repositories"
	si "github.com/eure/si2018-server-side/restapi/summerintern"
	"github.com/go-openapi/runtime/middleware"
)

func GetMatches(p si.GetMatchesParams) middleware.Responder {
	userTokenEnt, err := repositories.NewUserTokenRepository().GetByToken(p.Token)

	if err != nil {
		return getMatchesInternalServerErrorRespose()
	}

	if userTokenEnt == nil {
		return getMatchesUnauthorizedResponse()
	}

	// int64になっているのでcastする必要がある
	limit := int(p.Limit)
	offset := int(p.Offset)
	userID := userTokenEnt.UserID

	userMatchRepository := repositories.NewUserMatchRepository()
	userMatches, err := userMatchRepository.FindByUserIDWithLimitOffset(userID, limit, offset)

	if err != nil {
		return getMatchesInternalServerErrorRespose()
	}

	var ids []int64
	var matchUserResponses entities.MatchUserResponses

	for _, userMatch := range userMatches {
		ids = append(ids, userMatch.PartnerID)

		userMatchResponse := entities.MatchUserResponse{MatchedAt: userMatch.CreatedAt}
		matchUserResponses = append(matchUserResponses, userMatchResponse)
	}

	var users entities.Users
	users, err = repositories.NewUserRepository().FindByIDs(ids)
	if err != nil {
		return getMatchesInternalServerErrorRespose()
	}

	userImageRepository := repositories.NewUserImageRepository()
	images, err := userImageRepository.GetByUserIDs(ids)
	if err != nil {
		return getMatchesInternalServerErrorRespose()
	}

	for i := range users {
		users[i].ImageURI = images[i].Path
	}

	matchUserResponses = matchUserResponses.ApplyUsers(users)
	matchUser := matchUserResponses.Build()

	return si.NewGetMatchesOK().WithPayload(matchUser)
}
