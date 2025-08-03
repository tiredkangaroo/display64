export interface Provider {
  name: string;
  authorized: boolean;
  authorization_url: string;
  is_current: boolean;
}