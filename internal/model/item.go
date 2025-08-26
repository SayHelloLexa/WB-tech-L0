package domain

type Item struct {
	ChrtID      int    `json:"chrt_id"`
    TrackNumber string `json:"track_number"`
    Price       float64    `json:"price"`
    Rid         string `json:"rid"`
    Name        string `json:"name"`
    Sale        int    `json:"sale"`
    Size        string `json:"size"`
    TotalPrice  float64    `json:"total_price"`
    NmID        int    `json:"nm_id"`
    Brand       string `json:"brand"`
    Status      int    `json:"status"`
}