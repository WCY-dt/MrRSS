<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import { store } from '../../store.js';
import { PhPlus, PhTrash, PhFunnel, PhProhibit } from "@phosphor-icons/vue";

const props = defineProps({
    show: { type: Boolean, default: false },
    currentFilters: { type: Array, default: () => [] }
});

const emit = defineEmits(['close', 'apply']);

// Filter conditions
const conditions = ref([]);

// Field options
const fieldOptions = [
    { value: 'feed_name', labelKey: 'feedName', multiSelect: true },
    { value: 'feed_category', labelKey: 'feedCategory', multiSelect: true },
    { value: 'article_title', labelKey: 'articleTitle', multiSelect: false },
    { value: 'published_after', labelKey: 'publishedAfter', multiSelect: false },
    { value: 'published_before', labelKey: 'publishedBefore', multiSelect: false },
    { value: 'is_read', labelKey: 'readStatus', multiSelect: false, booleanField: true },
    { value: 'is_favorite', labelKey: 'favoriteStatus', multiSelect: false, booleanField: true }
];

// Operator options for article title only
const textOperatorOptions = [
    { value: 'contains', labelKey: 'contains' },
    { value: 'exact', labelKey: 'exactMatch' }
];

// Boolean value options for read/favorite status
const booleanOptions = [
    { value: 'true', labelKey: 'yes' },
    { value: 'false', labelKey: 'no' }
];

// Logic options (only AND/OR for connectors)
const logicOptions = [
    { value: 'and', labelKey: 'and' },
    { value: 'or', labelKey: 'or' }
];

// Feed names for multi-select
const feedNames = computed(() => {
    return store.feeds.map(f => f.title);
});

// Feed categories for multi-select
const feedCategories = computed(() => {
    const categories = new Set();
    store.feeds.forEach(f => {
        if (f.category) {
            categories.add(f.category);
        }
    });
    return Array.from(categories);
});

// Watch for modal show changes to reload filters
watch(() => props.show, (newVal) => {
    if (newVal && props.currentFilters && props.currentFilters.length > 0) {
        conditions.value = JSON.parse(JSON.stringify(props.currentFilters));
    }
});

onMounted(() => {
    // Load existing filters if provided
    if (props.currentFilters && props.currentFilters.length > 0) {
        conditions.value = JSON.parse(JSON.stringify(props.currentFilters));
    }
});

function addCondition() {
    conditions.value.push({
        id: Date.now(),
        logic: conditions.value.length > 0 ? 'and' : null,
        negate: false,  // NOT is now a modifier on the condition itself
        field: 'article_title',
        operator: 'contains',
        value: '',
        values: []  // For multi-select fields
    });
}

function removeCondition(index) {
    conditions.value.splice(index, 1);
    // Reset first condition's logic to null
    if (conditions.value.length > 0 && index === 0) {
        conditions.value[0].logic = null;
    }
}

function isDateField(field) {
    return field === 'published_after' || field === 'published_before';
}

function isMultiSelectField(field) {
    return field === 'feed_name' || field === 'feed_category';
}

function isBooleanField(field) {
    return field === 'is_read' || field === 'is_favorite';
}

function needsOperator(field) {
    // Only article_title needs the contains/exact operator
    return field === 'article_title';
}

function onFieldChange(index) {
    const condition = conditions.value[index];
    if (isDateField(condition.field)) {
        condition.operator = null;
        condition.value = '';
        condition.values = [];
    } else if (isMultiSelectField(condition.field)) {
        condition.operator = 'contains';  // Always contains for multi-select
        condition.value = '';
        condition.values = [];
    } else if (isBooleanField(condition.field)) {
        condition.operator = null;
        condition.value = 'true';  // Default to true (read/favorited)
        condition.values = [];
    } else {
        condition.operator = 'contains';
        condition.value = '';
        condition.values = [];
    }
}

function toggleNegate(index) {
    conditions.value[index].negate = !conditions.value[index].negate;
}

// Track which dropdown is open
const openDropdownIndex = ref(null);

function toggleDropdown(index) {
    if (openDropdownIndex.value === index) {
        openDropdownIndex.value = null;
    } else {
        openDropdownIndex.value = index;
    }
}

function toggleMultiSelectValue(index, val) {
    const condition = conditions.value[index];
    const idx = condition.values.indexOf(val);
    if (idx > -1) {
        condition.values.splice(idx, 1);
    } else {
        condition.values.push(val);
    }
}

// Get display text for multi-select values
function getMultiSelectDisplayText(condition, labelKey) {
    if (!condition.values || condition.values.length === 0) {
        return store.i18n.t(labelKey);
    }
    
    if (condition.values.length === 1) {
        return condition.values[0];
    }
    
    // Show first item and total count
    // For Chinese: "xxx等N个" means "xxx and N items in total"
    // For English: "xxx and N more" means N additional items
    const firstItem = condition.values[0];
    const totalCount = condition.values.length;
    const remaining = totalCount - 1;
    
    // Use different count based on locale
    const locale = store.i18n.locale.value;
    if (locale === 'zh') {
        return `${firstItem} ${store.i18n.t('andNMore', { count: totalCount })}`;
    }
    return `${firstItem} ${store.i18n.t('andNMore', { count: remaining })}`;
}

function clearFilters() {
    conditions.value = [];
    openDropdownIndex.value = null;
    // Auto-apply when clearing filters
    emit('apply', []);
    emit('close');
}

function applyFilters() {
    // Validate conditions - include conditions with value or values array
    const validConditions = conditions.value.filter(c => {
        if (isMultiSelectField(c.field)) {
            return c.values && c.values.length > 0;
        }
        return c.value;
    });
    
    emit('apply', validConditions);
    emit('close');
}

function close() {
    emit('close');
}
</script>

<template>
    <div v-if="show" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-bg-primary w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in">
            <!-- Header -->
            <div class="p-4 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
                <h3 class="text-lg font-semibold m-0 flex items-center gap-2">
                    <PhFunnel :size="20" />
                    {{ store.i18n.t('filterArticles') }}
                </h3>
                <span @click="close" class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary">&times;</span>
            </div>
            
            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-4 sm:p-6">
                <!-- Empty state -->
                <div v-if="conditions.length === 0" class="text-center text-text-secondary py-8">
                    <PhFunnel :size="48" class="mx-auto mb-3 opacity-50" />
                    <p>{{ store.i18n.t('noFiltersApplied') }}</p>
                </div>
                
                <!-- Condition list -->
                <div v-else class="space-y-3">
                    <div v-for="(condition, index) in conditions" :key="condition.id">
                        <!-- Logic connector (AND/OR) between conditions - styled distinctly -->
                        <div v-if="index > 0" class="flex items-center justify-center my-3">
                            <div class="flex-1 h-px bg-border"></div>
                            <div class="logic-connector mx-3">
                                <button 
                                    v-for="opt in logicOptions" 
                                    :key="opt.value"
                                    @click="condition.logic = opt.value"
                                    :class="['logic-btn', condition.logic === opt.value ? 'active' : '']">
                                    {{ store.i18n.t(opt.labelKey) }}
                                </button>
                            </div>
                            <div class="flex-1 h-px bg-border"></div>
                        </div>
                        
                        <!-- Condition card -->
                        <div class="condition-row bg-bg-secondary border border-border rounded-lg p-3 sm:p-4">
                            <div class="flex flex-wrap gap-2 items-end">
                                <!-- NOT toggle button -->
                                <div class="flex-shrink-0">
                                    <label class="block text-xs text-text-secondary mb-1">&nbsp;</label>
                                    <button 
                                        @click="toggleNegate(index)" 
                                        :class="['not-btn', condition.negate ? 'active' : '']"
                                        :title="store.i18n.t('not')">
                                        <PhProhibit :size="16" />
                                        <span class="text-xs font-medium">{{ store.i18n.t('not') }}</span>
                                    </button>
                                </div>
                                
                                <!-- Field selector -->
                                <div class="flex-1 min-w-[130px]">
                                    <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterField') }}</label>
                                    <select v-model="condition.field" @change="onFieldChange(index)" class="select-field w-full">
                                        <option v-for="opt in fieldOptions" :key="opt.value" :value="opt.value">
                                            {{ store.i18n.t(opt.labelKey) }}
                                        </option>
                                    </select>
                                </div>
                                
                                <!-- Operator selector (only for article_title) -->
                                <div v-if="needsOperator(condition.field)" class="w-28">
                                    <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterOperator') }}</label>
                                    <select v-model="condition.operator" class="select-field w-full">
                                        <option v-for="opt in textOperatorOptions" :key="opt.value" :value="opt.value">
                                            {{ store.i18n.t(opt.labelKey) }}
                                        </option>
                                    </select>
                                </div>
                                
                                <!-- Value input -->
                                <div class="flex-1 min-w-[140px]">
                                    <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterValue') }}</label>
                                    
                                    <!-- Date input for date fields -->
                                    <input v-if="isDateField(condition.field)" 
                                           type="date" 
                                           v-model="condition.value" 
                                           class="date-field w-full">
                                    
                                    <!-- Boolean select for read/favorite status -->
                                    <select v-else-if="isBooleanField(condition.field)"
                                            v-model="condition.value"
                                            class="select-field w-full">
                                        <option v-for="opt in booleanOptions" :key="opt.value" :value="opt.value">
                                            {{ store.i18n.t(opt.labelKey) }}
                                        </option>
                                    </select>
                                    
                                    <!-- Multi-select dropdown for feed name -->
                                    <div v-else-if="condition.field === 'feed_name'" class="dropdown-container">
                                        <button type="button" 
                                                @click="toggleDropdown(index)"
                                                class="dropdown-trigger">
                                            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText(condition, 'feedName') }}</span>
                                            <span class="dropdown-arrow">▼</span>
                                        </button>
                                        <div v-if="openDropdownIndex === index" 
                                             :class="['dropdown-menu', index === 0 ? 'dropdown-down' : 'dropdown-up']" 
                                             role="listbox" 
                                             :aria-label="store.i18n.t('feedName')">
                                            <div v-for="name in feedNames" :key="name" 
                                                 @click.stop="toggleMultiSelectValue(index, name)"
                                                 role="option"
                                                 :aria-selected="condition.values.includes(name)"
                                                 :class="['dropdown-option', condition.values.includes(name) ? 'selected' : '']">
                                                <input type="checkbox" 
                                                       :checked="condition.values.includes(name)" 
                                                       class="checkbox-input"
                                                       :aria-label="name"
                                                       tabindex="-1">
                                                <span class="truncate">{{ name }}</span>
                                            </div>
                                            <div v-if="feedNames.length === 0" class="text-text-secondary text-sm p-2">
                                                {{ store.i18n.t('noArticles') }}
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <!-- Multi-select dropdown for category -->
                                    <div v-else-if="condition.field === 'feed_category'" class="dropdown-container">
                                        <button type="button" 
                                                @click="toggleDropdown(index)"
                                                class="dropdown-trigger">
                                            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText(condition, 'feedCategory') }}</span>
                                            <span class="dropdown-arrow">▼</span>
                                        </button>
                                        <div v-if="openDropdownIndex === index" 
                                             :class="['dropdown-menu', index === 0 ? 'dropdown-down' : 'dropdown-up']" 
                                             role="listbox" 
                                             :aria-label="store.i18n.t('feedCategory')">
                                            <div v-for="cat in feedCategories" :key="cat" 
                                                 @click.stop="toggleMultiSelectValue(index, cat)"
                                                 role="option"
                                                 :aria-selected="condition.values.includes(cat)"
                                                 :class="['dropdown-option', condition.values.includes(cat) ? 'selected' : '']">
                                                <input type="checkbox" 
                                                       :checked="condition.values.includes(cat)" 
                                                       class="checkbox-input"
                                                       :aria-label="cat"
                                                       tabindex="-1">
                                                <span class="truncate">{{ cat }}</span>
                                            </div>
                                            <div v-if="feedCategories.length === 0" class="text-text-secondary text-sm p-2">
                                                {{ store.i18n.t('noArticles') }}
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <!-- Regular text input for article title -->
                                    <input v-else 
                                           type="text" 
                                           v-model="condition.value" 
                                           class="input-field w-full"
                                           :placeholder="store.i18n.t('filterValue')">
                                </div>
                                
                                <!-- Remove button -->
                                <div class="flex-shrink-0">
                                    <label class="block text-xs text-text-secondary mb-1">&nbsp;</label>
                                    <button @click="removeCondition(index)" class="btn-danger-icon" :title="store.i18n.t('removeCondition')">
                                        <PhTrash :size="18" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Add condition button -->
                <button @click="addCondition" class="btn-secondary w-full mt-4 flex items-center justify-center gap-2">
                    <PhPlus :size="18" />
                    {{ store.i18n.t('addCondition') }}
                </button>
            </div>
            
            <!-- Footer -->
            <div class="p-4 sm:p-5 border-t border-border bg-bg-secondary flex justify-between gap-3 shrink-0">
                <button @click="clearFilters" class="btn-secondary" :disabled="conditions.length === 0">
                    {{ store.i18n.t('clearFilters') }}
                </button>
                <button @click="applyFilters" class="btn-primary">
                    {{ store.i18n.t('applyFilters') }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
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
.btn-danger-icon {
    @apply p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors cursor-pointer;
}

/* Logic connector styling - distinct visual appearance */
.logic-connector {
    @apply flex items-center gap-1 bg-bg-tertiary rounded-full p-1;
}
.logic-btn {
    @apply px-3 py-1 text-xs font-bold rounded-full transition-all cursor-pointer;
    @apply text-text-secondary bg-transparent;
}
.logic-btn:hover {
    @apply text-text-primary bg-bg-secondary;
}
.logic-btn.active {
    @apply text-white bg-accent;
}

/* NOT button styling */
.not-btn {
    @apply flex items-center gap-1 px-2 py-2 rounded-md border transition-all cursor-pointer;
    @apply text-text-secondary bg-bg-primary border-border;
}
.not-btn:hover {
    @apply border-red-400 text-red-500;
}
.not-btn.active {
    @apply bg-red-500/10 border-red-500 text-red-500;
}

/* Dropdown multi-select styling */
.dropdown-container {
    @apply relative;
}
.dropdown-trigger {
    @apply w-full p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm;
    @apply flex items-center justify-between cursor-pointer hover:border-accent transition-colors;
}
.dropdown-text {
    @apply flex-1 text-left;
}
.dropdown-arrow {
    @apply text-text-secondary text-xs ml-2;
}
.dropdown-menu {
    @apply absolute left-0 right-0 border border-border rounded-md bg-bg-primary;
    @apply max-h-40 overflow-y-auto z-50 shadow-lg;
}
.dropdown-menu.dropdown-up {
    @apply bottom-full mb-1;
}
.dropdown-menu.dropdown-down {
    @apply top-full mt-1;
}
.dropdown-option {
    @apply flex items-center gap-2 px-3 py-2 cursor-pointer text-sm text-text-primary hover:bg-bg-tertiary;
}
.dropdown-option.selected {
    background-color: rgba(59, 130, 246, 0.1);
}
.checkbox-input {
    @apply w-4 h-4 accent-accent cursor-pointer;
}

.animate-fade-in {
    animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes modalFadeIn {
    from { transform: translateY(-20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}
</style>
