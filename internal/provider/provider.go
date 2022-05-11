package provider

type Hosts struct {
	List   []string
	active int
}

func NewProvider() Hosts {
	return Hosts{
		List:   make([]string, 0),
		active: 0,
	}
}

func (h *Hosts) Add(host string) {
	h.List = append(h.List, host)
}

func (h *Hosts) GetHost() string {
	host := h.List[h.active]
	h.active = (h.active + 1) % len(h.List)

	return host
}
