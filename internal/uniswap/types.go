package uniswap

type Bundles struct {
	Data struct {
		Bundles []struct {
			EthPrice string `json:"ethPrice"`
		} `json:"bundles"`
	} `json:"data"`
}

type Tokens struct {
	Data struct {
		Tokens []struct {
			Id             string `json:"id"`
			Name           string `json:"name"`
			Symbol         string `json:"symbol"`
			DerivedETH     string `json:"derivedETH"`
			TotalLiquidity string `json:"totalLiquidity"`
		} `json:"tokens"`
	} `json:"data"`
}
