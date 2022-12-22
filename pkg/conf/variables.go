package conf

const (
	Vendor      = "cisco"
	Version     = "0.0.2"
	App         = "argome"
	FeatureName = "tfc-agent-feature"
	Cookie      = "AuthCookie=eyJhbGciOiJSUzI1NiIsImtpZCI6InhqbWRlOGZueW5sdTczdDYwcHhvNWNybWEwMW8zMWdhIiwidHlwIjoiSldUIn0.eyJhdnBhaXIiOiJzaGVsbDpkb21haW5zPWFsbC9hZG1pbi8iLCJjbHVzdGVyIjoiNjI2NTJkNzAtNmY2NC0zMTJkLTZlNjQtMzIwMDAwMDAwMDAwIiwiZXhwIjoxNjM4NDY0MzI2LCJpYXQiOjE2Mzg0NjMxMjYsImlkIjoiNDhkMTA1YmRmYmM0OWE1ZmNmMzlhMTBiOTYxMzg2ZTYxZGZlNDAwODVjYjAzMTVkODE4Yjc2MWM1NzM1ZGFmYSIsImlzcyI6Im5kIiwiaXNzLWhvc3QiOiIxMC4yMy4yNDguNjUiLCJyYmFjIjpbeyJkb21haW4iOiJhbGwiLCJyb2xlcyI6W1siYWRtaW4iLCJXcml0ZVByaXYiXSxbImFwcC11c2VyIiwiUmVhZFByaXYiXV0sInJvbGVzUiI6MTY3NzcyMTYsInJvbGVzVyI6MX1dLCJzZXNzaW9uaWQiOiJQMXdDVVFGWEZCd1U0QTR5bjNZWVVmRmkiLCJ1c2VyZmxhZ3MiOjAsInVzZXJpZCI6MjUwMDIsInVzZXJuYW1lIjoiYWRtaW4iLCJ1c2VydHlwZSI6ImxvY2FsIn0.ONnDREp2wwbgfzxYEVWxbMMUqYFVkad26j63CCdM7AqYyNLVqu0g52QM10KQrbXhId1K5Rw9KXG7ZXYM-Swhw__4z15OiUTsC2UgHA6wPnBwr-uOUKonp53jU2Ok7Ae0xTlXPmRkdzyZuh8k6-c7D6sQJMtmOTj5N60GWisOInOam3VjIhPP1oLJMpAJSpRegno6aZGRM0NH9SvTVdRu1ZjlFa5rEtYydNl-2D00NhQ3yOMv_4ToAyqQLB9TmBuJKFa0q2DxQFfPjm__IyRtEnQ1rrdoxQ_dwxN-e8xeRL2BrZu_viWPLZ1mQw-QEiTWXtaa2RIBwM548UCQKeJn_A"
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
