/**
 * RSSHub integration types
 */

export interface RouteMatch {
  route: string;
  title: string;
  docs?: string;
  params: Record<string, string>;
  score: number;
}

export interface RSSHubMatchError extends Error {
  code: 'RSSHUB_MATCH_FOUND';
  original_url: string;
  matches: RouteMatch[];
}

export interface RSSHubSettings {
  rsshub_enabled: boolean;
  rsshub_instance_url: string;
  rsshub_fallback_enabled: boolean;
  rsshub_api_key: string;
}

export interface RadarRule {
  title: string;
  docs: string;
  source: string[];
  target: string;
}
