package errs

const (
	CODE_OK = 0

	// General error codes.
	CODE_ServerInternalError = 500  // Server internal error
	CODE_ArgsError           = 1001 // Input parameter error
	CODE_NoPermissionError   = 1002 // Insufficient permission
	CODE_DuplicateKeyError   = 1003
	CODE_RecordNotFoundError = 1004 // Record does not exist
	CODE_Unmarshal           = 1005
	CODE_Marshal             = 1006

	CODE_TokenExpiredError     = 1501
	CODE_TokenInvalidError     = 1502
	CODE_TokenMalformedError   = 1503
	CODE_TokenNotValidYetError = 1504
	CODE_TokenUnknownError     = 1505
	CODE_TokenKickedError      = 1506
	CODE_TokenNotExistError    = 1507
)
