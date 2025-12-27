/**
 * Cypress Test Setup Script
 *
 * This script prepares the environment for E2E testing by:
 * 1. Setting up a test database
 * 2. Seeding test data
 * 3. Ensuring the backend is ready
 */

import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

const TEST_DB_PATH = path.join(process.cwd(), 'test-data.db');

interface SetupOptions {
  clean?: boolean;
  seed?: boolean;
}

/**
 * Check if backend is running
 */
function checkBackend(): boolean {
  try {
    const response = execSync('curl -s http://localhost:34115/api/health', {
      stdio: 'ignore',
      timeout: 2000,
    });
    return true;
  } catch {
    return false;
  }
}

/**
 * Start the Wails development server
 */
function startBackend(): void {
  console.log('🚀 Starting Wails backend...');

  const isWindows = process.platform === 'win32';
  const startCommand = isWindows
    ? 'start /B wails3 dev > logs/backend.log 2>&1'
    : 'wails3 dev > logs/backend.log 2>&1 &';

  try {
    execSync(startCommand, { stdio: 'inherit' });
    console.log('✅ Backend started');
  } catch (error) {
    console.error('❌ Failed to start backend:', error);
    throw error;
  }
}

/**
 * Clean test database
 */
function cleanDatabase(): void {
  console.log('🧹 Cleaning test database...');

  if (fs.existsSync(TEST_DB_PATH)) {
    fs.unlinkSync(TEST_DB_PATH);
    console.log('✅ Test database cleaned');
  }
}

/**
 * Create test database schema
 */
function createTestSchema(): void {
  console.log('📊 Creating test database schema...');

  // This would typically use your migration system
  // For now, we'll rely on the backend auto-creating it
  console.log('✅ Test schema will be created by backend');
}

/**
 * Seed test data
 */
async function seedTestData(): Promise<void> {
  console.log('🌱 Seeding test data...');

  // Example test feeds
  const testFeeds = [
    {
      title: 'Test Feed 1',
      feed_url: 'https://example.com/feed1.xml',
      site_url: 'https://example.com',
      category: 'Tech',
    },
    {
      title: 'Test Feed 2',
      feed_url: 'https://example.com/feed2.xml',
      site_url: 'https://example.com',
      category: 'News',
    },
  ];

  try {
    // Seed feeds via API
    for (const feed of testFeeds) {
      execSync(
        `curl -X POST http://localhost:34115/api/feeds -H "Content-Type: application/json" -d '${JSON.stringify(feed)}'`,
        { stdio: 'ignore' }
      );
    }
    console.log('✅ Test data seeded');
  } catch (error) {
    console.warn('⚠️  Failed to seed test data (backend may not be ready yet)');
  }
}

/**
 * Wait for backend to be ready
 */
async function waitForBackend(maxAttempts = 30): Promise<void> {
  console.log('⏳ Waiting for backend to be ready...');

  for (let i = 0; i < maxAttempts; i++) {
    if (checkBackend()) {
      console.log('✅ Backend is ready');
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  throw new Error('Backend failed to start within expected time');
}

/**
 * Main setup function
 */
export async function setup(options: SetupOptions = {}): Promise<void> {
  console.log('\n🔧 Setting up Cypress test environment...\n');

  // Create logs directory
  if (!fs.existsSync(path.join(process.cwd(), 'logs'))) {
    fs.mkdirSync(path.join(process.cwd(), 'logs'), { recursive: true });
  }

  // Check if backend is already running
  if (!checkBackend()) {
    startBackend();
    await waitForBackend();
  } else {
    console.log('✅ Backend is already running');
  }

  // Clean database if requested
  if (options.clean) {
    cleanDatabase();
  }

  // Create schema
  createTestSchema();

  // Seed test data if requested
  if (options.seed) {
    await seedTestData();
  }

  console.log('\n✨ Setup complete! Tests can now run.\n');
}

// CLI interface
if (import.meta.url === `file://${process.argv[1]}`) {
  const args = process.argv.slice(2);
  const options: SetupOptions = {
    clean: args.includes('--clean'),
    seed: args.includes('--seed'),
  };

  setup(options).catch((error) => {
    console.error('\n❌ Setup failed:', error.message);
    process.exit(1);
  });
}
