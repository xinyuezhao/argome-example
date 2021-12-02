package conf

const (
	Vendor      = "cisco"
	Version     = "0.0.2"
	App         = "argome"
	FeatureName = "tfc-agent-feature"
	Cookie      = "AuthCookie=eyJhbGciOiJSUzI1NiIsImtpZCI6InhqbWRlOGZueW5sdTczdDYwcHhvNWNybWEwMW8zMWdhIiwidHlwIjoiSldUIn0.eyJhdnBhaXIiOiJzaGVsbDpkb21haW5zPWFsbC9hZG1pbi8iLCJjbHVzdGVyIjoiNjI2NTJkNzAtNmY2NC0zMTJkLTZlNjQtMzIwMDAwMDAwMDAwIiwiZXhwIjoxNjM4NDYzMTU2LCJpYXQiOjE2Mzg0NjE5NTYsImlkIjoiNDhkMTA1YmRmYmM0OWE1ZmNmMzlhMTBiOTYxMzg2ZTYxZGZlNDAwODVjYjAzMTVkODE4Yjc2MWM1NzM1ZGFmYSIsImlzcyI6Im5kIiwiaXNzLWhvc3QiOiIxMC4yMy4yNDguNjUiLCJyYmFjIjpbeyJkb21haW4iOiJhbGwiLCJyb2xlcyI6W1siYWRtaW4iLCJXcml0ZVByaXYiXSxbImFwcC11c2VyIiwiUmVhZFByaXYiXV0sInJvbGVzUiI6MTY3NzcyMTYsInJvbGVzVyI6MX1dLCJzZXNzaW9uaWQiOiJQMXdDVVFGWEZCd1U0QTR5bjNZWVVmRmkiLCJ1c2VyZmxhZ3MiOjAsInVzZXJpZCI6MjUwMDIsInVzZXJuYW1lIjoiYWRtaW4iLCJ1c2VydHlwZSI6ImxvY2FsIn0.SLDPWURWI2ppr7bW2UiU1QIBBca6SmHk76FHlpU3wjww5x9uk0mPBKiXswOAfckK3arT1i7WiSPPxm2Kscr8e0cJOdiIndp1Y2miKnRuepCtGasGQ_j6FrObPIJigj_fwtUTT2_i4Sg0jc3qK6MA8u_FeyRZVZBTC90W4UO_d1XOSJNJU62PdvPF9GTxHifF5tbv6hTxh3kKANTR2sGI00MO73YKd1VXige9zC_aZbRnro8BmrMCM2zgVmCfaUKq_CCo3CTXhh_fobmn2NECvXkQ2BhuiZ8ONoHkGWBy9g6sqfRs1BL-D0NO0MSQ5SADyg-5X3OdI1H-D2Gm32bsRg"
	Usertoken   = "ai1yMKOzv3Mptg.atlasv1.lOseEHJzlB49Vz0fXTlFUFRGGTuugiP3040sr1MGGOkHgRqzQ9FrpiUJzyTH1DzzFTM"
)

type Feature struct {
	Instances []struct {
		Features []struct {
			Name             string
			Instance         string
			OperState        string
			ConfigParameters struct {
				Name  string
				Token string
			}
		}
	}
}

type Agents struct {
	Data []Agent
}

type Agent struct {
	Id         string
	Attributes struct {
		Name    string
		Status  string
		Version string
	}
}

type AgentStatus struct {
	Data Agent
}
