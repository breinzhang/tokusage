package domain

type Project struct {
	ID          string `json:"project_id"`
	Name        string `json:"project_name"`
	PathNorm    string `json:"project_path_norm"`
	PathDisplay string `json:"project_path_display"`
}
