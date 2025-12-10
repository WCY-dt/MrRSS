import { ref, type Ref } from 'vue';
import type { Article } from '@/types/models';

interface LabelSettings {
  enabled: boolean;
  provider: string;
  maxCount: number;
}

export function useArticleLabels() {
  const labelSettings = ref<LabelSettings>({
    enabled: false,
    provider: 'local',
    maxCount: 5,
  });
  const labelingArticles: Ref<Set<number>> = ref(new Set());
  let observer: IntersectionObserver | null = null;

  // Load label settings
  async function loadLabelSettings(): Promise<void> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      labelSettings.value = {
        enabled: data.label_enabled === 'true',
        provider: data.label_provider || 'local',
        maxCount: parseInt(data.label_max_count || '5'),
      };
    } catch (e) {
      console.error('Error loading label settings:', e);
    }
  }

  // Setup intersection observer for auto-labeling
  function setupIntersectionObserver(listRef: HTMLElement | null, articles: Article[]): void {
    if (observer) {
      observer.disconnect();
    }

    observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const articleId = parseInt((entry.target as HTMLElement).dataset.articleId || '0');
            const article = articles.find((a) => a.id === articleId);

            // Only label if article exists, has no labels, and is not already being labeled
            if (article && !hasLabels(article.labels) && !labelingArticles.value.has(articleId)) {
              labelArticle(article);
            }
          }
        });
      },
      {
        root: listRef,
        rootMargin: '100px',
        threshold: 0.1,
      }
    );
  }

  // Check if article has labels
  function hasLabels(labelsJson: string | undefined): boolean {
    if (!labelsJson) return false;
    try {
      const parsed = JSON.parse(labelsJson);
      return Array.isArray(parsed) && parsed.length > 0;
    } catch {
      return false;
    }
  }

  // Label an article
  async function labelArticle(article: Article): Promise<void> {
    if (labelingArticles.value.has(article.id)) return;

    labelingArticles.value.add(article.id);

    try {
      const res = await fetch('/api/label/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.id,
        }),
      });

      if (res.ok) {
        const data = await res.json();
        // Update the article in the store
        article.labels = JSON.stringify(data.labels || []);
      } else {
        const errorData = await res.json();
        if (errorData.error === 'missing_ai_api_key') {
          console.error('AI API key is required for AI-based labeling');
        } else {
          console.error('Error labeling article:', res.status);
        }
      }
    } catch (e) {
      console.error('Error labeling article:', e);
    } finally {
      labelingArticles.value.delete(article.id);
    }
  }

  // Observe an article element
  function observeArticle(el: Element | null): void {
    if (el && observer && labelSettings.value.enabled) {
      observer.observe(el);
    }
  }

  // Update label settings from event
  function handleLabelSettingsChange(enabled: boolean, provider: string, maxCount: number): void {
    labelSettings.value = { enabled, provider, maxCount };

    // Disconnect observer if labeling is disabled
    if (!enabled && observer) {
      observer.disconnect();
      observer = null;
    }
    // Re-observe if labeling is enabled
    else if (enabled && observer) {
      setTimeout(() => {
        const cards = document.querySelectorAll('[data-article-id]');
        cards.forEach((card) => observer?.observe(card));
      }, 100);
    }
  }

  // Parse labels from JSON string
  function parseLabels(labelsJson: string | undefined): string[] {
    if (!labelsJson) return [];
    try {
      const parsed = JSON.parse(labelsJson);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }

  // Cleanup
  function cleanup(): void {
    if (observer) {
      observer.disconnect();
      observer = null;
    }
  }

  return {
    labelSettings,
    labelingArticles,
    loadLabelSettings,
    setupIntersectionObserver,
    labelArticle,
    observeArticle,
    handleLabelSettingsChange,
    parseLabels,
    cleanup,
  };
}
