import { computed } from 'vue';
import { PhRobot } from '@phosphor-icons/vue';

// AI provider detection patterns - based on model name patterns
const AI_PROVIDERS: Record<
  string,
  { name: string; icon: string; patterns: string[]; fallbackIcon: any }
> = {
  openai: {
    name: 'OpenAI',
    icon: '/assets/ai_icons/openai.svg',
    patterns: ['gpt', 'chatgpt', 'o1-'],
    fallbackIcon: PhRobot,
  },
  claude: {
    name: 'Claude',
    icon: '/assets/ai_icons/claude.svg',
    patterns: ['claude'],
    fallbackIcon: PhRobot,
  },
  gemini: {
    name: 'Gemini',
    icon: '/assets/ai_icons/gemini.svg',
    patterns: ['gemini', 'gemini-'],
    fallbackIcon: PhRobot,
  },
  deepseek: {
    name: 'DeepSeek',
    icon: '/assets/ai_icons/deepseek.svg',
    patterns: ['deepseek', 'deepseek-'],
    fallbackIcon: PhRobot,
  },
  zhipu: {
    name: 'Zhipu',
    icon: '/assets/ai_icons/zhipu.svg',
    patterns: ['glm', 'zhipu', 'chatglm'],
    fallbackIcon: PhRobot,
  },
  qwen: {
    name: 'Qwen',
    icon: '/assets/ai_icons/qwen.svg',
    patterns: ['qwen', 'qwq', 'qw-'],
    fallbackIcon: PhRobot,
  },
  wenxin: {
    name: 'Wenxin',
    icon: '/assets/ai_icons/wenxin.svg',
    patterns: ['ernie', 'wenxin'],
    fallbackIcon: PhRobot,
  },
  yuanbao: {
    name: 'Yuanbao',
    icon: '/assets/ai_icons/yuanbao.svg',
    patterns: ['hunyuan', 'yuanbao'],
    fallbackIcon: PhRobot,
  },
  minimax: {
    name: 'MiniMax',
    icon: '/assets/ai_icons/minimax.svg',
    patterns: ['abab', 'minimax'],
    fallbackIcon: PhRobot,
  },
  meta: {
    name: 'Meta',
    icon: '/assets/ai_icons/meta.svg',
    patterns: ['llama', 'llama-', 'meta-'],
    fallbackIcon: PhRobot,
  },
  grok: {
    name: 'Grok',
    icon: '/assets/ai_icons/grok.svg',
    patterns: ['grok'],
    fallbackIcon: PhRobot,
  },
};

// Detect AI provider from model name
function detectAIProvider(modelName: string): string | null {
  const model = modelName.toLowerCase();

  // Check for common patterns
  for (const [providerId, provider] of Object.entries(AI_PROVIDERS)) {
    for (const pattern of provider.patterns) {
      if (model.includes(pattern)) {
        return providerId;
      }
    }
  }

  return null;
}

// Get provider icon URL
export function getProviderIconUrl(modelName: string): string | null {
  const provider = detectAIProvider(modelName);
  return provider ? AI_PROVIDERS[provider]?.icon : null;
}
