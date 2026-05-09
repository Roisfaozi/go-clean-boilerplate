// Re-export everything from schemas for backward compatibility
export * from './schemas';

export interface Resource {
  id: string;
  name: string;
  slug: string;
  description?: string;
  status?: string;
  created_at?: number;
  updated_at?: number;
}

export interface Endpoint {
  id: string;
  name: string;
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  resource_id: string;
  resource_name?: string;
  description?: string;
  auth_required?: boolean;
  status?: string;
  created_at?: number;
  updated_at?: number;
}
