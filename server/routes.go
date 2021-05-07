package server

// setupRoutes sets up routes for gin application
func (s *Server) setupRoutes() {
	s.engine.GET("/health", s.handlers.Health.Handle)
	// swagger:route GET /status getStatus
	//
	// Gets latest status
	//
	// This will show latest status of the chain/node and indexer
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: ChainDetailsView
	//       400: BadRequestResponse
	s.engine.GET("/status", s.handlers.GetStatus.Handle)
	// swagger:route GET /block getBlockByHeight
	//
	// Gets block details and extrinsics for given height
	//
	// If no height is provided, will return last indexed block
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: BlockDetailsView
	//       400: BadRequestResponse
	s.engine.GET("/block", s.handlers.GetBlockByHeight.Handle)
	// swagger:route GET /block_times/:limit getBlockTimes
	//
	// Gets stats related to how fast blocks are being produced in the chain
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: BlockTimesView
	//       400: BadRequestResponse
	s.engine.GET("/block_times/:limit", s.handlers.GetBlockTimes.Handle)
	// swagger:route GET /blocks_summary getBlocksSummary
	//
	// Gets blocks summary
	//
	// Gets summary for blocks
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: BlockSummary
	//       400: BadRequestResponse
	s.engine.GET("/blocks_summary", s.handlers.GetBlockSummary.Handle)
	// swagger:route GET /transactions getTransactions
	//
	// Gets all signed transactions for height
	//
	// This will show all signed transactions for height. If height not specified, returns transactions
	// for most recently synced height.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: TransactionsView
	//       400: BadRequestResponse
	s.engine.GET("/transactions", s.handlers.GetTransactionsByHeight.Handle)
	// swagger:route GET /account_details/:stash_account getAccountDetails
	//
	// Gets latest account details
	//
	// This will show latest account balances, latest identity associated with the account (display name, email, social medias, etc), and all balance related events.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: AccountDetailsView
	//       400: BadRequestResponse
	s.engine.GET("/account_details/:stash_account", s.handlers.GetAccountDetails.Handle)
	// swagger:route GET /account_rewards/:stash_account getAccountRewards
	//
	// Gets rewards for account for time period
	//
	// This will show rewards for account for given time period from "start" to "end", If "end" is not specified,
	// will return rewards until most recently indexed block
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: AccountRewardsView
	//       400: BadRequestResponse
	s.engine.GET("/account_rewards/:stash_account", s.handlers.GetAccountRewards.Handle)
	// swagger:route GET /account/:stash_account getAccountByHeight
	//
	// Gets account balance details for given height
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: AccountHeightDetailsView
	//       400: BadRequestResponse
	s.engine.GET("/account/:stash_account", s.handlers.GetAccountByHeight.Handle)
	// swagger:route GET /system_events/:address getSystemEventsForAddress
	//
	// Gets system events for an address
	//
	// Returns lists of system events of given "kind" if specified in query (eg."active_balance_change_1", "commission_change_1", left_set", "joined_set", "missed_n_consecutive", etc. )
	// after given block height if provided. Otherwise returns all system events for account.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: SystemEventsView
	//       400: BadRequestResponse
	s.engine.GET("/system_events/:address", s.handlers.GetSystemEventsForAddress.Handle)
	// swagger:route GET /validator/:stash_account getValidatorByStash
	//
	// Gets aggregate details for validator
	//
	// Returns most recent aggreate information for validator account, as well as lists of validator era sequences for last eras_limit eras,
	// validator sessio sequences for last sessions_limit sessions, and most recent delegations.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: ValidatorAggregateView
	//       400: BadRequestResponse
	s.engine.GET("/validator/:stash_account", s.handlers.GetValidatorByStashAccount.Handle)
	// swagger:route GET /validators/for_min_height/:height getValidatorForMinHeight
	//
	// Gets all validator aggregates for validators which are active after min height
	//
	// Returns list of most recent aggreate information for all validator accounts that have been validators after provided min height
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: ValidatorAggregatesView
	//       400: BadRequestResponse
	s.engine.GET("/validators/for_min_height/:height", s.handlers.GetValidatorsForMinHeight.Handle)
	// swagger:route GET /validators getValidatorsByHeight
	//
	// Gets all validators for height
	//
	// Returns list of session and era validators for given height. If no height is provided, returns validators from most recently indexed height.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: ValidatorsView
	//       400: BadRequestResponse
	s.engine.GET("/validators", s.handlers.GetValidatorsByHeight.Handle)
	// swagger:route GET /validators_summary getValidatorSummary
	//
	// Gets all validators for height
	//
	// Returns validator summaries for all accounts (or single account if stash_account query is provided)
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: ValidatorsSummariesView
	//       400: BadRequestResponse
	s.engine.GET("/validators_summary", s.handlers.GetValidatorSummary.Handle)
	// swagger:route GET /rewards/:stash_account getRewards
	//
	// Gets rewards for account for eras
	//
	// This will show all rewards (claimed and unclaimed) for account for given eras from "start" to "end" eras. If "end" is not specified,
	// will return rewards until most recently indexed block.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: RewardsForErasView
	//       400: BadRequestResponse
	s.engine.GET("/rewards/:stash_account", s.handlers.GetRewardsForStashAccount.Handle)
	// swagger:route GET /apr/:stash_account getAPR
	//
	// Gets apr for account
	//
	// This will show daily reward aprs for an account for given time period from "start" to "end". If "end" is not specified,
	// will return aprs until most recently indexed block.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: RewardsForErasView
	//       400: BadRequestResponse
	s.engine.GET("/apr/:stash_account", s.handlers.GetAPRByAddress.Handle)
}
