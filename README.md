# กินดี — Budget based Food Recommendation

AI assisted food discovery built with React + TypeScript and a Go API. The API applies hard eligibility constraints (budget, selected diet, allergens, and optional sodium cap), then ranks eligible meals using a linear model whose weights can be trained in Colab and exported to Go. The filters are explicit user requirements; ranking is model driven rather than randomized or a hand written recommendation branch.

## Project layout

- `web/`: React/Vite interface.
- `api/recommend.go`: Vercel Go Function (`POST /api/recommend`).
- `internal/recommender/`: shared catalog, model artifact, filtering and inference.
- `notebooks/train_ranker.ipynb`: Colab workflow to retrieve USDA FoodData Central foods, train/evaluate a weakly supervised ranking proxy, and export Go weights.

## Run in VS Code

Requirements: Go 1.22+ and Node.js 20+.

1. Open this folder in VS Code.
2. Run `cd web && npm install` once.
3. In VS Code, choose **Terminal → Run Task → Foodwise: Run app**, or run `cd web && npm run dev`.
4. Open the Vite URL shown in terminal. The Vite dev server starts the Go API automatically and forwards `/api` requests to it.

Requires Go 1.22+ and Node.js 20+.

## Train in Google Colab

Open `notebooks/train_ranker.ipynb` in Colab. Create a free USDA FoodData Central API key and add it in Colab Secrets as `FDC_API_KEY`, or paste it when prompted. Run all cells. Download the generated `model.json` and replace `internal/recommender/model.json` in this project. The USDA sample CSV is saved in Colab's file pane; download it if you need to retain the training snapshot. Commit the model and redeploy.

USDA FoodData Central is the nutrition data source, not a Thai menu, price, or preference-rating dataset. The bundled Thai menu catalog is a small hand-curated demo with approximate prices and estimated nutrition. The notebook uses actual FDC nutrient observations to train a *weakly supervised proxy* whose declared target favors protein density and lower sodium; it uses sampled preference/budget context where FDC has no such fields. Reported metrics evaluate agreement with that proxy only, not real user satisfaction or clinical benefit. Replace proxy supervision with consented user feedback and validate against a user/time holdout before making preference-quality claims.

## Deploy to Vercel

Import this repository into Vercel, using this project folder as the root. `vercel.json` builds React and serves `api/recommend.go` as the Go function. Vercel currently documents Go Functions as Beta. Training is an offline Colab step; no model training occurs in a request. Use a database and verified local vendor data for production scale.

## API example

`POST /api/recommend`

```json
{"budget":80,"preferences":["spicy","plant"],"avoid":["peanut"],"diet":"any","max_sodium_mg":800,"target_calories":500,"limit":6}
```

Response contains ranked meals, reasons, and training status. The displayed match score is a ranking score, not a probability.

No medical advice is provided. Allergen metadata does not guarantee avoidance of cross contamination. Verify ingredients with vendors, especially for severe allergies. Nutrition and prices are indicative demo values.

## Sources

- USDA FoodData Central API and datasets: https://fdc.nal.usda.gov/api-guide/ and https://fdc.nal.usda.gov/download-datasets/
- Vercel Go Functions: https://vercel.com/docs/functions/runtimes/go
