package rustack

type Client struct {
	manager      *Manager
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	PaymentModel string  `json:"payment_model"`
	Balance      float32 `json:"contract.balance"`
}

func (m *Manager) GetClients(extraArgs ...Arguments) (clients []*Client, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/client"
	err = m.GetItems(path, args, &clients)
	for i := range clients {
		clients[i].manager = m
	}
	return
}
