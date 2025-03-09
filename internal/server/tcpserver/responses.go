package tcpserver

// Response messages
const (
	ResponseFailedReadCommand      = "Failed to read command\n"
	ResponseInvalidPowSolution     = "Invalid PoW solution.\n"
	ResponsePowVerificationSuccess = "PoW verification successful.\n"
	ResponseCommandTimeout         = "Timeout reading command.\n"
	QuoteServerError               = "Server error, please try again later.\n"
	ResponseGeneratedWisdom        = "Generated wisdom: %s\n"
	ResponseUnknownCommand         = "Unknown command\n"
)
