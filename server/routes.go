package server

// setupRoutes sets up routes for gin application
func (s *Server) setupRoutes() {
	s.engine.GET("/health", s.handlers.Health.Handle)
	s.engine.GET("/status", s.handlers.GetStatus.Handle)
	s.engine.GET("/block", s.handlers.GetBlockByHeight.Handle)
	s.engine.GET("/block_times/:limit", s.handlers.GetBlockTimes.Handle)
	s.engine.GET("/blocks_summary", s.handlers.GetBlockSummary.Handle)
	s.engine.GET("/transactions", s.handlers.GetTransactionsByHeight.Handle)
	//s.engine.GET("/validator/:address", s.handlers.GetValidatorByAddress.Handle)
	s.engine.GET("/validators/for_min_height/:height", s.handlers.GetValidatorsForMinHeight.Handle)
	s.engine.GET("/validators", s.handlers.GetValidatorsByHeight.Handle)
	s.engine.GET("/validators_summary", s.handlers.GetValidatorSummary.Handle)
	s.engine.GET("/staking", s.handlers.GetStakingDetailsByHeight.Handle)
	s.engine.GET("/delegations", s.handlers.GetDelegationsByHeight.Handle)
	s.engine.GET("/debonding_delegations", s.handlers.GetDebondingDelegationsByHeight.Handle)
	//s.engine.GET("/account/:address", s.handlers.GetAccountByAddress.Handle)
}
