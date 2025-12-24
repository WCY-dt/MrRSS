<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { RouteMatch } from '@/types/rsshub';

interface Props {
  visible: boolean;
  originalUrl: string;
  matches: RouteMatch[];
}

defineProps<Props>();

const emit = defineEmits<{
  close: [];
  select: [match: RouteMatch];
}>();

const { t } = useI18n();

function selectRoute(match: RouteMatch) {
  emit('select', match);
}

function close() {
  emit('close');
}

function getScoreColor(score: number): string {
  if (score >= 0.8) return 'text-green-500';
  if (score >= 0.5) return 'text-yellow-500';
  return 'text-gray-400';
}

function getScoreLabel(score: number): string {
  if (score >= 0.8) return t('rsshub.route_selector.high_match');
  if (score >= 0.5) return t('rsshub.route_selector.medium_match');
  return t('rsshub.route_selector.low_match');
}
</script>

<template>
  <Transition name="modal">
    <div
      v-if="visible"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
      data-modal-open="true"
      @click.self="close"
    >
      <div
        class="bg-bg-primary w-full max-w-2xl h-full sm:h-auto sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in"
      >
        <!-- Header -->
        <div class="p-3 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
          <div>
            <h3 class="text-base sm:text-lg font-semibold m-0">
              {{ t('rsshub.route_selector.title') }}
            </h3>
            <p class="text-xs text-text-secondary mt-1">
              {{ t('rsshub.route_selector.subtitle', { count: matches.length }) }}
            </p>
          </div>
          <span
            class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary"
            @click="close"
            >&times;</span
          >
        </div>

        <!-- Original URL -->
        <div class="px-4 sm:px-6 py-3 bg-bg-secondary border-b border-border">
          <label class="block text-xs font-semibold text-text-secondary mb-1">
            {{ t('rsshub.route_selector.original_url') }}
          </label>
          <div class="text-sm text-text-primary truncate font-mono">
            {{ originalUrl }}
          </div>
        </div>

        <!-- Route List -->
        <div class="flex-1 overflow-y-auto p-4 sm:p-6">
          <div class="space-y-3">
            <div
              v-for="match in matches"
              :key="match.route"
              class="group border border-border rounded-lg p-4 hover:border-accent hover:bg-bg-secondary cursor-pointer transition-all"
              :class="{ 'border-green-500/50 bg-green-500/5': match.score >= 0.8 }"
              @click="selectRoute(match)"
            >
              <!-- Header with title and score -->
              <div class="flex justify-between items-start mb-2">
                <h4 class="text-sm font-semibold text-text-primary m-0 flex-1">
                  {{ match.title }}
                </h4>
                <div class="flex items-center gap-2 ml-3">
                  <span
                    v-if="match.score >= 0.8"
                    class="text-xs px-2 py-0.5 rounded-full bg-green-500/20 text-green-600 dark:text-green-400"
                  >
                    {{ t('rsshub.route_selector.recommended') }}
                  </span>
                  <span :class="['text-xs', getScoreColor(match.score)]">
                    {{ getScoreLabel(match.score) }}
                  </span>
                </div>
              </div>

              <!-- Route path -->
              <div class="mb-2">
                <code class="text-xs bg-bg-tertiary px-2 py-1 rounded text-accent break-all">
                  {{ match.route }}
                </code>
              </div>

              <!-- Docs link -->
              <div v-if="match.docs" class="mt-2">
                <a
                  :href="match.docs"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-xs text-accent hover:underline inline-flex items-center gap-1"
                  @click.stop
                >
                  {{ t('rsshub.route_selector.view_docs') }}
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                    />
                  </svg>
                </a>
              </div>

              <!-- Hover indicator -->
              <div
                class="text-xs text-text-tertiary opacity-0 group-hover:opacity-100 transition-opacity mt-2"
              >
                {{ t('rsshub.route_selector.click_to_select') }}
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-3 sm:p-5 border-t border-border shrink-0">
          <button
            type="button"
            class="w-full px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary transition-colors"
            @click="close"
          >
            {{ t('common.cancel') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from > div,
.modal-leave-to > div {
  transform: scale(0.95);
}

.modal-enter-active > div,
.modal-leave-active > div {
  transition: transform 0.2s ease;
}
</style>
