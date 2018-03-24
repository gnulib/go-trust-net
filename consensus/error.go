package consensus

import (

)

const (
	ERR_NOT_IMPLEMENTED = 0x001
	ERR_INITIALIZATION_FAILED = 0x002
	ERR_DB_CORRUPTED = 0x003
	ERR_INVALID_ARG = 0x100
	ERR_TYPE_INCORRECT = 0x400
	ERR_TX_NOT_FOUND = 0x404
	ERR_STATE_INCORRECT = 0x407
	ERR_BLOCK_UNHASHED = 0x500
	ERR_DUPLICATE_TX = 0x600
	ERR_DUPLICATE_BLOCK = 0x601
	ERR_UPDATE_FAILED = 0x606
	ERR_BLOCK_VALIDATION = 0x700
	ERR_BLOCK_ORPHAN = 0x701
)