export const translations = {
    en: {
        // App
        appName: 'MrRSS',
        
        // Sidebar
        allArticles: 'All Articles',
        unread: 'Unread',
        favorites: 'Favorites',
        uncategorized: 'Uncategorized',
        addFeed: 'Add Feed',
        settings: 'Settings',
        
        // Article List
        articles: 'Articles',
        refresh: 'Refresh',
        search: 'Search...',
        noArticles: 'No articles found.',
        
        // Article Detail
        back: 'Back',
        markAsUnread: 'Mark as Unread',
        markAsRead: 'Mark as Read',
        toggleFavorite: 'Toggle Favorite',
        openInBrowser: 'Open in Browser',
        
        // Context Menu
        unsubscribe: 'Unsubscribe',
        editSubscription: 'Edit Subscription',
        renameCategory: 'Rename Category',
        addToFavorites: 'Add to Favorites',
        removeFromFavorites: 'Remove from Favorites',
        selectArticle: 'Select an article to start reading',
        
        // Settings Modal
        settingsTitle: 'Settings',
        general: 'General',
        feeds: 'Feeds',
        about: 'About',
        
        // General Settings
        appearance: 'Appearance',
        theme: 'Theme',
        themeDesc: 'Choose your preferred color scheme',
        light: 'Light',
        dark: 'Dark',
        auto: 'Auto (Follow System)',
        darkMode: 'Dark Mode',
        darkModeDesc: 'Switch between light and dark themes',
        updates: 'Updates',
        autoUpdateInterval: 'Auto-update Interval',
        autoUpdateIntervalDesc: 'How often to check for new articles (in minutes)',
        database: 'Database',
        autoCleanup: 'Auto Cleanup',
        autoCleanupDesc: 'Automatically remove old articles to save space',
        maxCacheSize: 'Max Cache Size',
        maxCacheSizeDesc: 'Maximum database size before cleanup (in MB)',
        maxArticleAge: 'Max Article Age',
        maxArticleAgeDesc: 'Delete articles older than this many days (except favorites)',
        days: 'days',
        translation: 'Translation',
        enableTranslation: 'Enable Translation',
        enableTranslationDesc: 'Automatically translate article titles to your preferred language',
        translationProvider: 'Translation Provider',
        deeplApiKey: 'DeepL API Key',
        deeplApiKeyPlaceholder: 'Enter your DeepL API key',
        targetLanguage: 'Target Language',
        language: 'Language',
        languageDesc: 'Select interface language',
        
        // Languages
        english: 'English',
        spanish: 'Spanish',
        french: 'French',
        german: 'German',
        chinese: 'Chinese',
        japanese: 'Japanese',
        
        // Feeds Settings
        dataManagement: 'Data Management',
        importOPML: 'Import OPML',
        exportOPML: 'Export OPML',
        cleanDatabase: 'Clean Database',
        cleanDatabaseDesc: 'Removes all articles except read and favorited ones. Old articles are also automatically cleaned if "Auto Cleanup" is enabled in General settings.',
        manageFeeds: 'Manage Feeds',
        deleteSelected: 'Delete Selected',
        moveSelected: 'Move Selected',
        selectAll: 'Select All',
        edit: 'Edit',
        delete: 'Delete',
        
        // About
        aboutApp: 'A simple, modern RSS reader.',
        version: 'Version',
        viewOnGitHub: 'View on GitHub',
        checkForUpdates: 'Check for Updates',
        checking: 'Checking...',
        upToDate: 'You are using the latest version',
        updateAvailable: 'Update available',
        currentVersion: 'Current version',
        latestVersion: 'Latest version',
        downloadUpdate: 'Download Update',
        downloading: 'Downloading...',
        downloadComplete: 'Download complete',
        installingUpdate: 'Installing update...',
        updateWillRestart: 'The application will restart to install the update',
        downloadFailed: 'Download failed',
        installFailed: 'Installation failed',
        releaseNotes: 'Release Notes',
        errorCheckingUpdates: 'Error checking for updates',
        
        // Modals
        addNewFeed: 'Add New Feed',
        editFeed: 'Edit Feed',
        rssUrl: 'RSS URL',
        rssUrlPlaceholder: 'https://example.com/rss',
        category: 'Category',
        categoryOptional: 'Category (Optional)',
        categoryPlaceholder: 'e.g. Tech/News',
        title: 'Title',
        titlePlaceholder: 'Custom feed title',
        optional: 'Optional',
        addSubscription: 'Add Subscription',
        saveChanges: 'Save Changes',
        adding: 'Adding...',
        saving: 'Saving...',
        saveSettings: 'Save Settings',
        
        // Confirm Dialogs
        confirm: 'Confirm',
        cancel: 'Cancel',
        deleteFeedTitle: 'Delete Feed',
        deleteFeedMessage: 'Are you sure you want to delete this feed?',
        deleteMultipleFeedsTitle: 'Delete Multiple Feeds',
        deleteMultipleFeedsMessage: 'Are you sure you want to delete {count} feeds?',
        unsubscribeTitle: 'Unsubscribe',
        unsubscribeMessage: 'Are you sure you want to unsubscribe from {name}?',
        cleanDatabaseTitle: 'Clean Database',
        cleanDatabaseMessage: 'This will delete all articles except read and favorited ones. Continue?',
        clean: 'Clean',
        
        // Toast Messages
        feedAddedSuccess: 'Feed added successfully',
        feedUpdatedSuccess: 'Feed updated successfully',
        feedDeletedSuccess: 'Feed deleted successfully',
        feedsDeletedSuccess: 'Feeds deleted successfully',
        feedsMovedSuccess: 'Feeds moved successfully',
        unsubscribedSuccess: 'Successfully unsubscribed',
        databaseCleanedSuccess: 'Database cleaned up successfully. {count} articles deleted.',
        opmlImportedSuccess: 'OPML Imported. Feeds will appear shortly.',
        errorSavingSettings: 'Error saving settings',
        errorAddingFeed: 'Error adding feed',
        errorUpdatingFeed: 'Error updating feed',
        errorCleaningDatabase: 'Error cleaning up database',
        importFailed: 'Import failed: {error}',
        
        // Prompts
        enterCategoryName: 'Enter new category name:',
    },
    zh: {
        // App
        appName: 'MrRSS',
        
        // Sidebar
        allArticles: '所有文章',
        unread: '未读',
        favorites: '收藏',
        uncategorized: '未分类',
        addFeed: '添加订阅',
        settings: '设置',
        
        // Article List
        articles: '文章',
        refresh: '刷新',
        search: '搜索...',
        noArticles: '未找到文章。',
        
        // Article Detail
        back: '返回',
        markAsUnread: '标记为未读',
        markAsRead: '标记为已读',
        toggleFavorite: '切换收藏',
        openInBrowser: '在浏览器中打开',
        
        // Context Menu
        unsubscribe: '取消订阅',
        editSubscription: '编辑订阅',
        renameCategory: '重命名分类',
        addToFavorites: '添加到收藏',
        removeFromFavorites: '从收藏中移除',
        selectArticle: '选择一篇文章开始阅读',
        
        // Settings Modal
        settingsTitle: '设置',
        general: '常规',
        feeds: '订阅源',
        about: '关于',
        
        // General Settings
        appearance: '外观',
        theme: '主题',
        themeDesc: '选择您喜欢的配色方案',
        light: '亮色',
        dark: '暗色',
        auto: '自动（跟随系统）',
        darkMode: '暗色模式',
        darkModeDesc: '在亮色和暗色主题之间切换',
        updates: '更新',
        autoUpdateInterval: '自动更新间隔',
        autoUpdateIntervalDesc: '检查新文章的频率（分钟）',
        database: '数据库',
        autoCleanup: '自动清理',
        autoCleanupDesc: '自动删除旧文章以节省空间',
        maxCacheSize: '最大缓存大小',
        maxCacheSizeDesc: '清理前的最大数据库大小（MB）',
        maxArticleAge: '最大文章保留天数',
        maxArticleAgeDesc: '删除超过此天数的旧文章（收藏文章除外）',
        days: '天',
        translation: '翻译',
        enableTranslation: '启用翻译',
        enableTranslationDesc: '自动将文章标题翻译为您的首选语言',
        translationProvider: '翻译提供商',
        deeplApiKey: 'DeepL API 密钥',
        deeplApiKeyPlaceholder: '输入您的 DeepL API 密钥',
        targetLanguage: '目标语言',
        language: '语言',
        languageDesc: '选择界面语言',
        
        // Languages
        english: '英语',
        spanish: '西班牙语',
        french: '法语',
        german: '德语',
        chinese: '中文',
        japanese: '日语',
        
        // Feeds Settings
        dataManagement: '数据管理',
        importOPML: '导入 OPML',
        exportOPML: '导出 OPML',
        cleanDatabase: '清理数据库',
        cleanDatabaseDesc: '删除除已读和已收藏外的所有文章。如果在常规设置中启用了"自动清理"，旧文章也会被自动清理。',
        manageFeeds: '管理订阅源',
        deleteSelected: '删除选中',
        moveSelected: '移动选中',
        selectAll: '全选',
        edit: '编辑',
        delete: '删除',
        
        // About
        aboutApp: '一个简洁、现代的 RSS 阅读器。',
        version: '版本',
        viewOnGitHub: '在 GitHub 上查看',
        checkForUpdates: '检查更新',
        checking: '检查中...',
        upToDate: '您正在使用最新版本',
        updateAvailable: '有可用更新',
        currentVersion: '当前版本',
        latestVersion: '最新版本',
        downloadUpdate: '下载更新',
        downloading: '下载中...',
        downloadComplete: '下载完成',
        installingUpdate: '正在安装更新...',
        updateWillRestart: '应用程序将重启以安装更新',
        downloadFailed: '下载失败',
        installFailed: '安装失败',
        releaseNotes: '发行说明',
        errorCheckingUpdates: '检查更新时出错',
        
        // Modals
        addNewFeed: '添加新订阅',
        editFeed: '编辑订阅',
        rssUrl: 'RSS 地址',
        rssUrlPlaceholder: 'https://example.com/rss',
        category: '分类',
        categoryOptional: '分类（可选）',
        categoryPlaceholder: '例如：科技/新闻',
        title: '标题',
        titlePlaceholder: '自定义订阅标题',
        optional: '可选',
        addSubscription: '添加订阅',
        saveChanges: '保存更改',
        adding: '添加中...',
        saving: '保存中...',
        saveSettings: '保存设置',
        
        // Confirm Dialogs
        confirm: '确认',
        cancel: '取消',
        deleteFeedTitle: '删除订阅',
        deleteFeedMessage: '确定要删除这个订阅吗？',
        deleteMultipleFeedsTitle: '删除多个订阅',
        deleteMultipleFeedsMessage: '确定要删除 {count} 个订阅吗？',
        unsubscribeTitle: '取消订阅',
        unsubscribeMessage: '确定要取消订阅 {name} 吗？',
        cleanDatabaseTitle: '清理数据库',
        cleanDatabaseMessage: '这将删除除已读和已收藏外的所有文章。是否继续？',
        clean: '清理',
        
        // Toast Messages
        feedAddedSuccess: '订阅添加成功',
        feedUpdatedSuccess: '订阅更新成功',
        feedDeletedSuccess: '订阅删除成功',
        feedsDeletedSuccess: '订阅删除成功',
        feedsMovedSuccess: '订阅移动成功',
        unsubscribedSuccess: '取消订阅成功',
        databaseCleanedSuccess: '数据库清理成功。已删除 {count} 篇文章。',
        opmlImportedSuccess: 'OPML 已导入。订阅源即将显示。',
        errorSavingSettings: '保存设置时出错',
        errorAddingFeed: '添加订阅时出错',
        errorUpdatingFeed: '更新订阅时出错',
        errorCleaningDatabase: '清理数据库时出错',
        importFailed: '导入失败：{error}',
        
        // Prompts
        enterCategoryName: '输入新的分类名称：',
    }
};

export function createI18n() {
    const locale = ref(localStorage.getItem('locale') || 'en');
    
    const t = (key, params = {}) => {
        let text = translations[locale.value]?.[key] || translations.en[key] || key;
        
        // Replace placeholders like {count}, {name}, {error}
        Object.keys(params).forEach(param => {
            text = text.replace(`{${param}}`, params[param]);
        });
        
        return text;
    };
    
    const setLocale = (newLocale) => {
        if (translations[newLocale]) {
            locale.value = newLocale;
            localStorage.setItem('locale', newLocale);
        }
    };
    
    return {
        locale,
        t,
        setLocale
    };
}

// For use in script setup
import { ref } from 'vue';

export const i18n = createI18n();
