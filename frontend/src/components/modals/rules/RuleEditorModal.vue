<script setup lang="ts">
import { ref, computed, watch, type Ref, type ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLightning, PhPlus, PhFunnel, PhListChecks } from '@phosphor-icons/vue';
import RuleLogicConnector from './RuleLogicConnector.vue';
import RuleAction from './RuleAction.vue';
import RuleConditionItem from './RuleConditionItem.vue';
import {
  useRuleOptions,
  type Condition,
  isMultiSelectField,
} from '@/composables/rules/useRuleOptions';
import { useRuleConditions } from '@/composables/rules/useRuleConditions';
import { useRuleActions } from '@/composables/rules/useRuleActions';
import { useModalClose } from '@/composables/ui/useModalClose';

const { t } = useI18n();

// Modal close handling
useModalClose(() => handleClose());

// Use composables
const { actionOptions } = useRuleOptions();

const {
  openDropdownIndex,
  addCondition: addConditionHelper,
  removeCondition: removeConditionHelper,
  onFieldChange,
  toggleNegate,
  toggleDropdown,
} = useRuleConditions();

const {
  addAction: addActionHelper,
  removeAction: removeActionHelper,
  updateAction: updateActionHelper,
} = useRuleActions(actionOptions);

interface Rule {
  id: number;
  name: string;
  enabled: boolean;
  conditions: Condition[];
  actions: string[];
}

interface Props {
  show?: boolean;
  rule?: Rule | null;
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  rule: null,
});

const emit = defineEmits<{
  close: [];
  save: [rule: Rule];
}>();

// Form data
const ruleName = ref('');
const conditions: Ref<Condition[]> = ref([]);
const actions: Ref<string[]> = ref([]);

// Initialize form when rule changes
watch(
  () => props.rule,
  (newRule) => {
    if (newRule) {
      ruleName.value = newRule.name || '';
      conditions.value = newRule.conditions ? JSON.parse(JSON.stringify(newRule.conditions)) : [];
      actions.value = newRule.actions ? [...newRule.actions] : [];
    } else {
      ruleName.value = '';
      conditions.value = [];
      actions.value = [];
    }
  },
  { immediate: true }
);

// Condition helpers
function addCondition(): void {
  addConditionHelper(conditions.value);
}

function removeCondition(index: number): void {
  removeConditionHelper(conditions.value, index);
}

function handleFieldChange(index: number): void {
  onFieldChange(conditions.value[index]);
}

function handleToggleNegate(index: number): void {
  toggleNegate(conditions.value[index]);
}

// Action helpers
function addAction(): void {
  addActionHelper(actions);
}

function removeAction(index: number): void {
  removeActionHelper(actions, index);
}

function updateAction(index: number, value: string): void {
  updateActionHelper(actions, index, value);
}

// Form validation
const isValid: ComputedRef<boolean> = computed(() => {
  return actions.value.length > 0;
});

// Save handler
function handleSave(): void {
  if (!isValid.value) {
    window.showToast(t('noActionsSelected'), 'warning');
    return;
  }

  const rule: Rule = {
    id: props.rule ? props.rule.id : Date.now(),
    name: ruleName.value || t('rules'),
    enabled: props.rule ? props.rule.enabled : true,
    conditions: conditions.value.filter((c) => {
      if (isMultiSelectField(c.field)) {
        return c.values && c.values.length > 0;
      }
      return c.value !== '';
    }),
    actions: [...actions.value],
  };

  emit('save', rule);
}

function handleClose(): void {
  openDropdownIndex.value = null;
  emit('close');
}
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 backdrop-blur-sm p-0 sm:p-4"
    data-modal-open="true"
    style="will-change: transform; transform: translateZ(0)"
    @click.self="handleClose"
  >
    <div
      class="bg-bg-primary w-full max-w-2xl h-full sm:h-auto sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border-0 sm:border border-border overflow-hidden animate-fade-in"
    >
      <!-- Header -->
      <div class="p-4 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
        <h3 class="text-lg font-semibold m-0 flex items-center gap-2">
          <PhLightning :size="20" />
          {{ rule ? t('editRule') : t('addRule') }}
        </h3>
        <span
          class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary"
          @click="handleClose"
          >&times;</span
        >
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        <!-- Rule Name -->
        <div class="space-y-2">
          <label class="block text-sm font-medium">{{ t('ruleName') }}</label>
          <input
            v-model="ruleName"
            type="text"
            :placeholder="t('ruleNamePlaceholder')"
            class="input-field w-full"
          />
        </div>

        <!-- Conditions Section -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="flex items-center gap-2 text-sm font-medium">
              <PhFunnel :size="16" />
              {{ t('ruleCondition') }}
            </label>
          </div>

          <!-- Empty state -->
          <div
            v-if="conditions.length === 0"
            class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border"
          >
            <p class="text-sm">{{ t('conditionAlways') }}</p>
          </div>

          <!-- Condition list -->
          <div v-else class="space-y-3">
            <div v-for="(condition, index) in conditions" :key="condition.id">
              <!-- Logic connector -->
              <RuleLogicConnector
                v-if="index > 0"
                :logic="condition.logic || 'and'"
                @update="(logic) => (condition.logic = logic)"
              />

              <!-- Condition card -->
              <RuleConditionItem
                :condition="condition"
                :index="index"
                :is-dropdown-open="openDropdownIndex === index"
                @update:field="
                  (value) => {
                    condition.field = value;
                    handleFieldChange(index);
                  }
                "
                @update:operator="(value) => (condition.operator = value)"
                @update:value="(value) => (condition.value = value)"
                @update:values="(values) => (condition.values = values)"
                @update:negate="handleToggleNegate(index)"
                @toggle-dropdown="toggleDropdown(index)"
                @remove="removeCondition(index)"
              />
            </div>
          </div>

          <!-- Add condition button -->
          <button
            class="btn-secondary w-full flex items-center justify-center gap-2"
            @click="addCondition"
          >
            <PhPlus :size="16" />
            {{ t('addCondition') }}
          </button>
        </div>

        <!-- Actions Section -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="flex items-center gap-2 text-sm font-medium">
              <PhListChecks :size="16" />
              {{ t('ruleActions') }}
            </label>
          </div>

          <!-- Empty state -->
          <div
            v-if="actions.length === 0"
            class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border"
          >
            <p class="text-sm">{{ t('noActionsSelected') }}</p>
          </div>

          <!-- Action list -->
          <div v-else class="space-y-2">
            <RuleAction
              v-for="(action, index) in actions"
              :key="index"
              :action="action"
              :index="index"
              :selected-actions="actions"
              :all-action-options="actionOptions"
              @update="(value) => updateAction(index, value)"
              @remove="removeAction(index)"
            />
          </div>

          <!-- Add action button -->
          <button
            class="btn-secondary w-full flex items-center justify-center gap-2"
            :disabled="actions.length >= actionOptions.length"
            @click="addAction"
          >
            <PhPlus :size="16" />
            {{ t('addAction') }}
          </button>
        </div>
      </div>

      <!-- Footer -->
      <div
        class="p-4 sm:p-5 border-t border-border bg-bg-secondary flex justify-end gap-3 shrink-0"
      >
        <button class="btn-secondary" @click="handleClose">
          {{ t('cancel') }}
        </button>
        <button class="btn-primary" :disabled="!isValid" @click="handleSave">
          {{ t('saveChanges') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors;
}
.select-field {
  @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors cursor-pointer;
}
.date-field {
  @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors cursor-pointer;
  color-scheme: light dark;
}
.btn-primary {
  @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}
.btn-secondary {
  @apply bg-bg-tertiary text-text-primary border border-border px-4 py-2.5 rounded-lg cursor-pointer font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
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
