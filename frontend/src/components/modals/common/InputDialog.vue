<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue';
import { useModalClose } from '@/composables/ui/useModalClose';

interface Props {
  title?: string;
  message?: string;
  placeholder?: string;
  defaultValue?: string;
  confirmText?: string;
  cancelText?: string;
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Input',
  message: '',
  placeholder: '',
  defaultValue: '',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
});

const emit = defineEmits<{
  confirm: [value: string];
  cancel: [];
  close: [];
}>();

const inputValue = ref(props.defaultValue);
const inputRef: Ref<HTMLInputElement | null> = ref(null);

// Modal close handling - ESC should cancel
useModalClose(() => handleCancel());

onMounted(() => {
  // Focus the input when dialog opens
  if (inputRef.value) {
    inputRef.value.focus();
    inputRef.value.select();
  }
});

function handleConfirm() {
  emit('confirm', inputValue.value);
  emit('close');
}

function handleCancel() {
  emit('cancel');
  emit('close');
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault();
    handleConfirm();
  }
  // ESC is now handled by useModalClose
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4"
    data-modal-open="true"
    style="will-change: transform; transform: translateZ(0)"
    @click.self="handleCancel"
  >
    <div
      class="bg-bg-primary max-w-md w-full mx-2 sm:mx-4 rounded-xl shadow-2xl border border-border overflow-hidden animate-fade-in"
    >
      <div class="p-3 sm:p-5 border-b border-border">
        <h3 class="text-base sm:text-lg font-semibold m-0">{{ title }}</h3>
      </div>

      <div class="p-3 sm:p-5">
        <p v-if="message" class="m-0 mb-2 sm:mb-3 text-text-primary text-sm sm:text-base">
          {{ message }}
        </p>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          :placeholder="placeholder"
          class="input-field w-full text-sm sm:text-base"
          @keydown="handleKeyDown"
        />
      </div>

      <div
        class="p-3 sm:p-5 border-t border-border bg-bg-secondary flex flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:gap-3"
      >
        <button class="btn-secondary text-sm sm:text-base" @click="handleCancel">
          {{ cancelText }}
        </button>
        <button class="btn-primary text-sm sm:text-base" @click="handleConfirm">
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply px-3 py-2 rounded-lg border border-border bg-bg-secondary text-text-primary;
  @apply focus:outline-none focus:ring-2 focus:ring-accent;
}

.btn-primary {
  @apply bg-accent text-white border-none px-4 sm:px-5 py-2 sm:py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors;
}

.btn-secondary {
  @apply bg-transparent border border-border text-text-primary px-4 sm:px-5 py-2 sm:py-2.5 rounded-lg cursor-pointer font-medium hover:bg-bg-tertiary transition-colors;
}

.animate-fade-in {
  animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalFadeIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>
