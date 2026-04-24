package finalphotos

func AllowsFinalEditingOrDeliveryGeneration(status string) bool {
	return status == "editing" || status == "delivered"
}
