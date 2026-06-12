package apperrors

const (
	CodeReadyDatabasePingFailed = "READY_GET_V1_HANDLER_DATABASE_PING_FAILED"
	MsgReadyDatabasePingFailed  = "Database is not reachable"

	CodeReadyRabbitMQPingFailed = "READY_GET_V1_HANDLER_RABBITMQ_PING_FAILED"
	MsgReadyRabbitMQPingFailed  = "Message broker is not reachable"

	CodeProductCreateInvalidBody = "PRODUCT_CREATE_V1_HANDLER_INVALID_BODY"
	MsgProductCreateInvalidBody  = "Invalid product payload"

	CodeProductGetNotFound = "PRODUCT_GET_V1_SERVICE_NOT_FOUND"
	MsgProductGetNotFound  = "Product not found"

	CodeCreativeRunCreateInvalidBody = "CREATIVE_RUN_CREATE_V1_HANDLER_INVALID_BODY"
	MsgCreativeRunCreateInvalidBody  = "Invalid creative run payload"

	CodeCreativeRunCreateProductNotFound = "CREATIVE_RUN_CREATE_V1_SERVICE_PRODUCT_NOT_FOUND"
	MsgCreativeRunCreateProductNotFound  = "Product not found"

	CodeCreativeRunGetNotFound = "CREATIVE_RUN_GET_V1_SERVICE_NOT_FOUND"
	MsgCreativeRunGetNotFound  = "Creative run not found"

	CodeCreativeRunStartNotFound = "CREATIVE_RUN_START_V1_SERVICE_NOT_FOUND"
	MsgCreativeRunStartNotFound  = "Creative run not found"

	CodeCreativeRunStartInvalidState = "CREATIVE_RUN_START_V1_SERVICE_INVALID_STATE"
	MsgCreativeRunStartInvalidState  = "Creative run cannot be started in its current state"

	CodeCreativeRunStepEditInvalid = "CREATIVE_RUN_STEP_PATCH_V1_SERVICE_INVALID"
	MsgCreativeRunStepEditInvalid  = "Step cannot be edited in its current state"

	CodeCreativeRunReprocessInvalid = "CREATIVE_RUN_REPROCESS_V1_SERVICE_INVALID"
	MsgCreativeRunReprocessInvalid  = "Run cannot be reprocessed in its current state"
)
