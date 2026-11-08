package recommender

import (
	"embed"
	"encoding/json"
	"math"
	"sort"
	"strings"
)

//go:embed catalog.json model.json
var assets embed.FS

type Meal struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Price       int      `json:"price_thb"`
	Kcal        float64  `json:"kcal"`
	Protein     float64  `json:"protein_g"`
	Carbs       float64  `json:"carbs_g"`
	Fat         float64  `json:"fat_g"`
	Sodium      float64  `json:"sodium_mg"`
	Diet        string   `json:"diet"`
	Tags        []string `json:"tags"`
	Allergens   []string `json:"allergens"`
	Image       string   `json:"image"`
	PhotoURL    string   `json:"photo_url"`
	PhotoCredit string   `json:"photo_credit"`
	PhotoPage   string   `json:"photo_page"`
	Score       float64  `json:"score,omitempty"`
	Reasons     []string `json:"reasons,omitempty"`
}
type Model struct {
	Version  string    `json:"version"`
	Features []string  `json:"features"`
	Weights  []float64 `json:"weights"`
	Bias     float64   `json:"bias"`
	Trained  bool      `json:"trained"`
	Note     string    `json:"note,omitempty"`
}
type Request struct {
	Budget         int      `json:"budget"`
	Preferences    []string `json:"preferences"`
	Avoid          []string `json:"avoid"`
	Diet           string   `json:"diet"`
	MaxSodium      int      `json:"max_sodium_mg"`
	TargetCalories int      `json:"target_calories"`
	Limit          int      `json:"limit"`
}
type Result struct {
	Items []Meal `json:"items"`
	Model Model  `json:"model"`
	Count int    `json:"count"`
}

func load(name string, dst any) { b, _ := assets.ReadFile(name); _ = json.Unmarshal(b, dst) }
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}
func clamp(x float64) float64 { return math.Max(0, math.Min(1, x)) }
func Recommend(q Request) Result {
	var meals []Meal
	var model Model
	load("catalog.json", &meals)
	load("model.json", &model)
	if q.Budget < 1 {
		q.Budget = 100
	}
	if q.Limit < 1 || q.Limit > 12 {
		q.Limit = 6
	}
	out := make([]Meal, 0)
	for _, m := range meals {
		if m.Price > q.Budget {
			continue
		}
		excluded := false
		for _, a := range q.Avoid {
			if contains(m.Allergens, a) {
				excluded = true
			}
		}
		if excluded {
			continue
		}
		if q.Diet != "" && q.Diet != "any" && q.Diet != "omnivore" && m.Diet != "vegan" && m.Diet != q.Diet {
			continue
		}
		if q.MaxSodium > 0 && m.Sodium > float64(q.MaxSodium) {
			continue
		}
		f := map[string]float64{"budget_fit": clamp(1 - float64(q.Budget-m.Price)/float64(q.Budget)*1.3), "protein": clamp(m.Protein / 35), "low_sodium": clamp(1 - m.Sodium/1800), "preference_match": 0, "calorie_fit": 0.6, "popularity": 0.5}
		for _, tag := range q.Preferences {
			matched := contains(m.Tags, tag)
			if tag == "protein" && m.Protein >= 25 {
				matched = true
			}
			if matched {
				f["preference_match"]++
			}
		}
		if len(q.Preferences) > 0 {
			f["preference_match"] /= float64(len(q.Preferences))
		}
		if q.TargetCalories > 0 {
			f["calorie_fit"] = clamp(1 - math.Abs(m.Kcal-float64(q.TargetCalories))/float64(q.TargetCalories))
		}
		score := model.Bias
		for i, n := range model.Features {
			if i < len(model.Weights) {
				score += model.Weights[i] * f[n]
			}
		}
		m.Score = math.Round((1/(1+math.Exp(-score)))*1000) / 10
		if m.Diet == "vegan" {
			m.Reasons = append(m.Reasons, "plant-based")
		}
		if m.Protein >= 25 {
			m.Reasons = append(m.Reasons, "protein-rich")
		}
		for _, t := range q.Preferences {
			if contains(m.Tags, t) {
				m.Reasons = append(m.Reasons, "matches "+t)
				break
			}
		}
		m.Reasons = append(m.Reasons, "within budget")
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Price < out[j].Price
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return Result{Items: out, Model: model, Count: len(out)}
}
