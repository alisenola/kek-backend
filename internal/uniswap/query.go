package uniswap

import (
	"fmt"
)

func QueryBundles() map[string]string {
	return map[string]string{
		"query": `
			query bundles {
				bundles(where: { id: "1" }) {
					ethPrice
				}
			}
		`,
	}
}

func QueryToken(address string) map[string]string {
	query := fmt.Sprintf(`
		query tokens {
			tokens(where: { id: "%s" }) {
				id
				name
				symbol
				derivedETH
				totalLiquidity
			}
		}
	`, address)
	return map[string]string{"query": query}
}
