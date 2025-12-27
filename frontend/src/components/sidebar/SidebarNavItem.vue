<script setup lang="ts">
import {
  PhListDashes,
  PhSquaresFour,
  PhCircle,
  PhStar,
  PhClockCountdown,
  PhImages,
} from '@phosphor-icons/vue';
import { computed } from 'vue';
import type { Component } from 'vue';

interface Props {
  label: string;
  isActive: boolean;
  icon: 'all' | 'unread' | 'favorites' | 'readLater' | 'imageGallery';
  unreadCount?: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  click: [];
}>();

const iconMap: Record<string, Component> = {
  all: PhListDashes,
  unread: PhCircle,
  favorites: PhStar,
  readLater: PhClockCountdown,
  imageGallery: PhImages,
};

// Use different icon for "all" when active
const currentIcon = computed(() => {
  if (props.icon === 'all' && props.isActive) {
    return PhSquaresFour;
  }
  return iconMap[props.icon];
});

const iconWeight = computed(() => {
  // Use fill weight for active state, except for "all" in inactive state
  return props.isActive ? 'fill' : 'regular';
});
</script>

<template>
  <button :class="['nav-item', isActive ? 'active' : '']" @click="emit('click')">
    <component :is="currentIcon" :size="20" :weight="iconWeight" />
    <span class="flex-1 text-left">{{ label }}</span>
    <span v-if="unreadCount && unreadCount > 0" class="unread-badge">{{ unreadCount }}</span>
  </button>
</template>

<style scoped>
@reference "../../style.css";

.nav-item {
  @apply flex items-center gap-2 sm:gap-3 w-full px-2 sm:px-3 py-2 sm:py-2.5 text-text-secondary rounded-lg font-medium transition-colors hover:bg-bg-tertiary hover:text-text-primary text-left text-sm sm:text-base;
}
.nav-item.active {
  @apply bg-bg-tertiary text-accent font-semibold;
}
.unread-badge {
  @apply text-[9px] sm:text-[10px] font-semibold rounded-full min-w-[14px] sm:min-w-[16px] h-[14px] sm:h-[16px] px-0.5 sm:px-1 flex items-center justify-center;
  background-color: rgba(120, 120, 120, 0.25);
  color: #444444;
}
</style>

<style>
.dark-mode .unread-badge {
  background-color: rgba(100, 100, 100, 0.6) !important;
  color: #f0f0f0 !important;
}
</style>
