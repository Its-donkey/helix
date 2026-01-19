---
layout: default
title: Content Classification Labels API
description: Content Classification Labels (CCL) are used to categorize streams based on their content. These labels help viewers make informed decisions about the content they watch.
---

## GetContentClassificationLabels

Get information about Twitch content classification labels.

**Requires:** No authentication required

```go
// Get labels in default locale (English)
resp, err := client.GetContentClassificationLabels(ctx, nil)

// Get labels in a specific locale
resp, err = client.GetContentClassificationLabels(ctx, &helix.GetContentClassificationLabelsParams{
    Locale: "en-US",
})

// Other supported locales: "bg-BG", "cs-CZ", "da-DK", "de-DE", "el-GR", "en-GB",
// "es-ES", "es-MX", "fi-FI", "fr-FR", "hu-HU", "it-IT", "ja-JP", "ko-KR",
// "nl-NL", "no-NO", "pl-PL", "pt-BR", "pt-PT", "ro-RO", "ru-RU", "sk-SK",
// "sv-SE", "th-TH", "tr-TR", "vi-VN", "zh-CN", "zh-TW"

for _, label := range resp.Data {
    fmt.Printf("Label ID: %s\n", label.ID)
    fmt.Printf("Name: %s\n", label.Name)
    fmt.Printf("Description: %s\n", label.Description)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "DrugsIntoxication",
      "name": "Drugs, Intoxication, or Excessive Tobacco Use",
      "description": "Excessive tobacco use or the consumption of alcohol or other substances is a focus of my content."
    },
    {
      "id": "SexualThemes",
      "name": "Sexual Themes",
      "description": "Content that focuses on sexualized physical attributes and activities, sexual topics, or experiences."
    },
    {
      "id": "ViolentGraphic",
      "name": "Violent and Graphic Depictions",
      "description": "Simulations and/or depictions of realistic violence, gore, extreme injury, or death."
    },
    {
      "id": "Gambling",
      "name": "Significant Gambling",
      "description": "Participating in online or in-person gambling, poker or fantasy sports, that involve the exchange of real money."
    },
    {
      "id": "ProfanityVulgarity",
      "name": "Significant Profanity or Vulgarity",
      "description": "Prolonged, and repeated use of obscenities, profanities, and vulgarities, especially as a regular part of speech."
    }
  ]
}
```

