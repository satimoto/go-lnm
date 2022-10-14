package tokenauthorization

import (
	"context"
	"log"
	"time"

	dbUtil "github.com/satimoto/go-datastore/pkg/util"
)

func (r *TokenAuthorizationResolver) StartTokenAuthorizationMonitor(authorizationID string) {
	ctx := context.Background()
	waitingSeconds := 0
	timeoutSeconds := 90

	tokenAuthorization, err := r.Repository.GetTokenAuthorizationByAuthorizationID(ctx, authorizationID)
	
	if err != nil {
		dbUtil.LogOnError("LSP136", "Error retrieving token authorization", err)
		log.Printf("LSP136: AuthorizationID=%v", authorizationID)
		return
	}

	user, err := r.SessionResolver.UserResolver.Repository.GetUserByTokenID(ctx, tokenAuthorization.TokenID)

	if err != nil {
		dbUtil.LogOnError("LSP137", "Error retrieving token authorization", err)
		log.Printf("LSP137: AuthorizationID=%v TokenID=%v", authorizationID, tokenAuthorization.TokenID)
		return
	}

	r.SendTokenAuthorizeNotification(user, tokenAuthorization)

	log.Printf("Authorization monitor timeout set for %v seconds: %v", timeoutSeconds, authorizationID)
	time.Sleep(time.Duration(timeoutSeconds) * time.Second)

monitorLoop:
	for {
		tokenAuthorization, err := r.Repository.GetTokenAuthorizationByAuthorizationID(ctx, authorizationID)

		if err != nil {
			dbUtil.LogOnError("LSP138", "Error retrieving token authorization", err)
			log.Printf("LSP136: AuthorizationID=%v", authorizationID)
			break monitorLoop
		}

		if tokenAuthorization.Authorized {
			// Session was authorized
			log.Printf("Authorized: %v", authorizationID)
			break monitorLoop
		}

		session, err := r.SessionResolver.Repository.GetSessionByAuthorizationID(ctx, authorizationID)

		if err == nil {
			// Cancel session the unauthorized session
			log.Printf("Stopped unauthorizated session %v", session.Uid)
			r.SessionResolver.StopSession(ctx, session)
			break monitorLoop
		}

		_, err = r.CdrRepository.GetCdrByAuthorizationID(ctx, authorizationID)

		if err == nil || waitingSeconds >= 300 {
			// Cdr received for the AuthorizationID or hit the waiting timeout
			break monitorLoop
		}

		waitingSeconds += 10 
		time.Sleep(10 * time.Second)
	}

	log.Printf("Authorization monitor ended after %v seconds: %v", timeoutSeconds + waitingSeconds, authorizationID)
}
