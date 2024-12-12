package payment_networks

type CustomBank struct {
	Name      string
	NetworkId string
	ChannelId string
}

const NgnBankChannelId = "fe8f4989-3bf6-41ca-9621-ffe2bc127569"

var customBanks = []CustomBank{
	{
		Name:      "Moniepoint MFB",
		NetworkId: "ALLBANK",
		ChannelId: NgnBankChannelId,
	},
}

func GetCustomBanks(channelId string) []CustomBank {
	var result []CustomBank
	for _, bank := range customBanks {
		if bank.ChannelId == channelId {
			result = append(result, bank)
		}
	}
	return result
}
