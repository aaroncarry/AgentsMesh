import { request, orgPath } from "./base";

export interface ResourceGrant {
  id: number;
  resource_type: string;
  resource_id: string;
  user_id: number;
  granted_by: number;
  created_at: string;
  user?: { id: number; email: string; username: string; name?: string };
  granted_by_user?: { id: number; email: string; username: string; name?: string };
}

const resourcePlural: Record<string, string> = {
  pod: "pods",
  runner: "runners",
  repository: "repositories",
};

function grantPath(resourceType: string, resourceId: string, suffix = "") {
  const plural = resourcePlural[resourceType] || `${resourceType}s`;
  return orgPath(`/${plural}/${resourceId}/grants${suffix}`);
}

export const grantApi = {
  list: (resourceType: string, resourceId: string) =>
    request<{ grants: ResourceGrant[] }>(grantPath(resourceType, resourceId)),

  grant: (resourceType: string, resourceId: string, userId: number) =>
    request<{ grant: ResourceGrant }>(grantPath(resourceType, resourceId), {
      method: "POST",
      body: { user_id: userId },
    }),

  revoke: (resourceType: string, resourceId: string, grantId: number) =>
    request<{ message: string }>(grantPath(resourceType, resourceId, `/${grantId}`), {
      method: "DELETE",
    }),
};
