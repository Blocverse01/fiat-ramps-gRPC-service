package payment_methods

type Channel struct {
	Country     string
	RampType    string
	AccountType string
	ID          string
}

var RecommendedChannels = []Channel{
	{
		Country:     "BW",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "1239227d-f33a-4d7a-8520-0a0cf41988fb",
	},
	{
		Country:     "BW",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "602060c6-b4af-49a4-b8fd-2fdd0dbf13ce",
	},
	{
		Country:     "CM",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "353b1e71-f6f1-48f5-bc11-26af091c62fd",
	},
	{
		Country:     "CM",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "79da4d6e-1c42-4aac-ae7d-422730528f96",
	},
	{
		Country:     "CG",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "5968d5d4-103b-4333-98e2-b400968c6927",
	},
	{
		Country:     "CG",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "6fef21a2-b8f3-4515-bef5-63597ea0307f",
	},
	{
		Country:     "GA",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "0cc05476-c25a-46c4-9c71-23c39ed87348",
	},
	{
		Country:     "GA",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "65a69867-81e2-44c7-a97b-fcaaf5f38359",
	},
	{
		Country:     "CI",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "2452c27c-7c49-442c-b6f8-cde03f03f9ba",
	},
	{
		Country:     "CI",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "33a82864-6460-43d7-9fc0-911f9bd8d50a",
	},
	{
		Country:     "KE",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "7c7833d8-2f26-430e-8675-7ff07a5caf0c",
	},
	{
		Country:     "KE",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "c2b2eeda-d4ca-49fd-ba21-0781ffa7714b",
	},
	{
		Country:     "KE",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "de274d59-46b3-49a8-8836-3c0cc5b73a4f",
	},
	{
		Country:     "KE",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "b7140c26-0c93-460b-85cf-bdb53af66ca0",
	},
	{
		Country:     "MW",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "b7b89a56-b0a1-479b-8514-5cf60a8ce975",
	},
	{
		Country:     "MW",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "88c80c44-23a8-4729-9e36-7be766effefb",
	},
	{
		Country:     "MW",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "158f471c-62ac-469f-9127-012b3bc648e1",
	},
	{
		Country:     "MW",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "aa11c087-d0c8-4f8d-b566-3625a7581a1f",
	},
	{
		Country:     "ML",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "63bcd780-8304-4df4-8453-1201381e8615",
	},
	{
		Country:     "ML",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "4da264c2-080b-4cb6-9401-da68ea93bcf5",
	},
	{
		Country:     "NG",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "af944f0c-ba70-47c7-86dc-1bad5a6ab4e4",
	},
	{
		Country:     "NG",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "fe8f4989-3bf6-41ca-9621-ffe2bc127569",
	},
	{
		Country:     "RW",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "b9034627-69a4-43fd-90e5-220738919436",
	},
	{
		Country:     "RW",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "71764ae2-6620-4e9d-9474-4c94877366b9",
	},
	{
		Country:     "RW",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "05ec29bc-a6c0-4045-8ef1-8701fa991d12",
	},
	{
		Country:     "ZA",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "7c6e1ef6-ac88-40f8-8a0d-d077d0aa7851",
	},
	{
		Country:     "ZA",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "81018280-e320-4c81-9b2f-6f636c2239d8",
	},
	{
		Country:     "TZ",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "fc7a5bc2-9100-473e-8c01-df563494ee73",
	},
	{
		Country:     "TZ",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "12b6d583-b5ad-440f-8185-02c67259b17d",
	},
	{
		Country:     "UG",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "5de46c23-6461-4950-871f-a971490772cf",
	},
	{
		Country:     "UG",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "48f72484-f2a5-4bef-8bc1-f57ccd1c837d",
	},
	{
		Country:     "ZM",
		RampType:    "deposit",
		AccountType: "bank",
		ID:          "f436afb5-d034-4d18-b9d5-e8125e6ddfe4",
	},
	{
		Country:     "ZM",
		RampType:    "withdraw",
		AccountType: "bank",
		ID:          "ddc30e71-58e4-4bb4-a22b-ef0293f176e4",
	},
	{
		Country:     "ZM",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "3137dba1-9856-4880-80c0-3ec43d50d3f6",
	},
	{
		Country:     "ZM",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "de392cbe-9826-45e1-aede-69ad2be6e3a0",
	},
	{
		Country:     "TG",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "603c84e5-d219-444b-a5bb-40f756291ebe",
	},
	{
		Country:     "TG",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "3955a441-fa81-4bee-a3ce-f958a530be80",
	},
	{
		Country:     "SN",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "50d03c41-2c8a-4a31-a2ca-d05a924d7816",
	},
	{
		Country:     "SN",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "030d5d5e-95a5-45b8-8340-d9d56bdb25f6",
	},
	{
		Country:     "BF",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "9c94cad1-ffc4-4211-ae69-e62c72827d01",
	},
	{
		Country:     "BF",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "fbd1276a-8709-44e0-ac1d-9b41656591cd",
	},
	{
		Country:     "BJ",
		RampType:    "deposit",
		AccountType: "phone",
		ID:          "13ba87e5-cfce-4380-90cd-76f41d05c419",
	},
	{
		Country:     "BJ",
		RampType:    "withdraw",
		AccountType: "phone",
		ID:          "1997a129-32f1-4aa7-8f2a-edaf3a0bb0a5",
	},
}

func GetRecommendedChannels(country, rampType string) []Channel {
	var result []Channel
	for _, channel := range RecommendedChannels {
		if channel.Country == country && channel.RampType == rampType {
			result = append(result, channel)
		}
	}
	return result
}
