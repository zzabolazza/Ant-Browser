package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const managedPrivacyExtensionDirName = "Facade Managed Privacy Extension"

type SpeechVoice struct {
	Name         string `json:"name"`
	VoiceURI     string `json:"voiceURI"`
	Lang         string `json:"lang"`
	LocalService bool   `json:"localService"`
	Default      bool   `json:"default"`
}

func EnsureManagedPrivacyExtension(userDataDir string, lang string) (string, error) {
	userDataDir = strings.TrimSpace(userDataDir)
	if userDataDir == "" {
		return "", fmt.Errorf("user data directory is empty")
	}
	extensionDir := filepath.Join(userDataDir, managedPrivacyExtensionDirName)
	if err := os.MkdirAll(extensionDir, 0o755); err != nil {
		return "", fmt.Errorf("create managed privacy extension directory: %w", err)
	}

	voices := speechVoicesForLanguage(lang)
	manifest := `{
  "manifest_version": 3,
  "name": "Facade Managed Privacy",
  "version": "1.0.0",
  "description": "Managed browser privacy shims for Facade profiles.",
  "content_scripts": [
    {
      "matches": ["<all_urls>"],
      "js": ["voices.js"],
      "run_at": "document_start",
      "all_frames": true,
      "world": "MAIN"
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(extensionDir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		return "", fmt.Errorf("write managed privacy extension manifest: %w", err)
	}

	script, err := buildVoicesShimScript(voices)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(extensionDir, "voices.js"), []byte(script), 0o644); err != nil {
		return "", fmt.Errorf("write managed privacy extension script: %w", err)
	}
	return extensionDir, nil
}

func buildVoicesShimScript(voices []SpeechVoice) (string, error) {
	payload, err := json.Marshal(voices)
	if err != nil {
		return "", fmt.Errorf("encode speech voices: %w", err)
	}
	return `(function () {
  if (!('speechSynthesis' in window)) return;
  var voiceData = ` + string(payload) + `;
  function makeVoice(item) {
    var voice = {
      name: item.name,
      voiceURI: item.voiceURI || item.name,
      lang: item.lang,
      localService: Boolean(item.localService),
      default: Boolean(item.default)
    };
    return Object.freeze(voice);
  }
  var voices = Object.freeze(voiceData.map(makeVoice));
  var synth = window.speechSynthesis;
  var handler = null;
  function getVoices() {
    return voices.slice();
  }
  try {
    Object.defineProperty(synth, 'getVoices', {
      configurable: true,
      enumerable: false,
      writable: false,
      value: getVoices
    });
  } catch (_) {
    try { synth.getVoices = getVoices; } catch (_) {}
  }
  try {
    Object.defineProperty(synth, 'onvoiceschanged', {
      configurable: true,
      enumerable: true,
      get: function () { return handler; },
      set: function (value) { handler = typeof value === 'function' ? value : null; }
    });
  } catch (_) {}
  function emitVoicesChanged() {
    try {
      if (handler) handler.call(synth, new Event('voiceschanged'));
    } catch (_) {}
    try {
      synth.dispatchEvent(new Event('voiceschanged'));
    } catch (_) {}
  }
  setTimeout(emitVoicesChanged, 0);
})();` + "\n", nil
}

func speechVoicesForLanguage(lang string) []SpeechVoice {
	lang = normalizeSpeechLang(lang)
	names := speechVoiceNames(lang)
	voices := make([]SpeechVoice, 0, len(names))
	for i, name := range names {
		voices = append(voices, SpeechVoice{
			Name:         name,
			VoiceURI:     name,
			Lang:         lang,
			LocalService: false,
			Default:      i == 0,
		})
	}
	return voices
}

func normalizeSpeechLang(lang string) string {
	value := strings.TrimSpace(lang)
	if value == "" {
		return "en-US"
	}
	value = strings.ReplaceAll(value, "_", "-")
	parts := strings.Split(value, "-")
	if len(parts) == 1 {
		switch strings.ToLower(parts[0]) {
		case "zh":
			return "zh-CN"
		case "ja":
			return "ja-JP"
		case "ko":
			return "ko-KR"
		case "de":
			return "de-DE"
		case "fr":
			return "fr-FR"
		case "es":
			return "es-ES"
		case "it":
			return "it-IT"
		default:
			return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[0])
		}
	}
	primary := strings.ToLower(parts[0])
	region := strings.ToUpper(parts[1])
	return primary + "-" + region
}

func speechVoiceNames(lang string) []string {
	switch strings.ToLower(lang) {
	case "en-sg":
		return []string{"Microsoft Luna Online (Natural) - English (Singapore)", "Google English (Singapore)"}
	case "en-us":
		return []string{"Microsoft Aria Online (Natural) - English (United States)", "Google US English"}
	case "en-gb":
		return []string{"Microsoft Sonia Online (Natural) - English (United Kingdom)", "Google UK English Female"}
	case "en-ca":
		return []string{"Microsoft Clara Online (Natural) - English (Canada)", "Google Canadian English"}
	case "en-au":
		return []string{"Microsoft Natasha Online (Natural) - English (Australia)", "Google Australian English"}
	case "zh-cn":
		return []string{"Microsoft Xiaoxiao Online (Natural) - Chinese (Mainland)", "Google Mandarin (China Mainland)"}
	case "zh-hk":
		return []string{"Microsoft HiuMaan Online (Natural) - Chinese (Hong Kong)", "Google Cantonese (Hong Kong)"}
	case "zh-tw":
		return []string{"Microsoft HsiaoChen Online (Natural) - Chinese (Taiwan)", "Google Mandarin (Taiwan)"}
	case "ja-jp":
		return []string{"Microsoft Nanami Online (Natural) - Japanese (Japan)", "Google 日本語"}
	case "ko-kr":
		return []string{"Microsoft SunHi Online (Natural) - Korean (Korea)", "Google Korean"}
	case "de-de":
		return []string{"Microsoft Katja Online (Natural) - German (Germany)", "Google Deutsch"}
	case "fr-fr":
		return []string{"Microsoft Denise Online (Natural) - French (France)", "Google French"}
	case "it-it":
		return []string{"Microsoft Elsa Online (Natural) - Italian (Italy)", "Google Italian"}
	case "es-es":
		return []string{"Microsoft Elvira Online (Natural) - Spanish (Spain)", "Google Spanish"}
	case "es-mx":
		return []string{"Microsoft Dalia Online (Natural) - Spanish (Mexico)", "Google Spanish (Mexico)"}
	default:
		return []string{"Microsoft Aria Online (Natural) - English (United States)", "Google US English"}
	}
}
