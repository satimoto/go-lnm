package user

import (
	"context"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *UserResolver) RestrictUser(ctx context.Context, user db.User) error {
	return r.updateUserRestriction(ctx, user, true, db.TokenAllowedTypeNOCREDIT, db.TokenWhitelistTypeNEVER)
}

func (r *UserResolver) UnrestrictUser(ctx context.Context, user db.User) error {
	return r.updateUserRestriction(ctx, user, false, db.TokenAllowedTypeALLOWED, db.TokenWhitelistTypeALLOWED)
}

func (r *UserResolver) updateUserRestriction(ctx context.Context, user db.User, restricted bool, tokenAllowed db.TokenAllowedType, tokenWhitelist db.TokenWhitelistType) error {
	if user.IsRestricted != restricted {
		updateUserParams := param.NewUpdateUserParams(user)
		updateUserParams.IsRestricted = restricted

		_, err := r.Repository.UpdateUser(ctx, updateUserParams)

		if err != nil {
			return err
		}

		updateTokensRequest := &ocpirpc.UpdateTokensRequest{
			UserId:    user.ID,
			Allowed:   string(tokenAllowed),
			Whitelist: string(tokenWhitelist),
		}

		_, err = r.OcpiService.UpdateTokens(ctx, updateTokensRequest)

		return err
	}

	return nil
}