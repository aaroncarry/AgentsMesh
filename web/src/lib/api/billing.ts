import { request } from "./base";

// Billing types
export interface SubscriptionPlan {
  id: number;
  name: string;
  display_name: string;
  price_per_seat_monthly: number;
  included_session_minutes: number;
  price_per_extra_minute: number;
  max_users: number;
  max_runners: number;
  max_repositories: number;
  features: Record<string, unknown>;
  is_active: boolean;
}

export interface UsageOverview {
  session_minutes: number;
  included_session_minutes: number;
  users: number;
  max_users: number;
  runners: number;
  max_runners: number;
  repositories: number;
  max_repositories: number;
}

export interface BillingOverview {
  plan: SubscriptionPlan;
  status: string;
  billing_cycle: string;
  current_period_start: string;
  current_period_end: string;
  usage: UsageOverview;
}

export interface Subscription {
  id: number;
  organization_id: number;
  plan_id: number;
  status: string;
  billing_cycle: string;
  current_period_start: string;
  current_period_end: string;
  plan?: SubscriptionPlan;
}

// Billing API
export const billingApi = {
  // Get billing overview
  getOverview: () =>
    request<{ overview: BillingOverview }>("/api/v1/org/billing/overview"),

  // Get subscription
  getSubscription: () =>
    request<{ subscription: Subscription }>("/api/v1/org/billing/subscription"),

  // Create subscription
  createSubscription: (planName: string, billingCycle?: string) =>
    request<{ subscription: Subscription }>("/api/v1/org/billing/subscription", {
      method: "POST",
      body: { plan_name: planName, billing_cycle: billingCycle || "monthly" },
    }),

  // Update subscription
  updateSubscription: (planName: string) =>
    request<{ subscription: Subscription }>("/api/v1/org/billing/subscription", {
      method: "PUT",
      body: { plan_name: planName },
    }),

  // Cancel subscription
  cancelSubscription: () =>
    request<{ message: string }>("/api/v1/org/billing/subscription", {
      method: "DELETE",
    }),

  // List available plans
  listPlans: () =>
    request<{ plans: SubscriptionPlan[] }>("/api/v1/org/billing/plans"),

  // Get usage
  getUsage: (type?: string) => {
    const params = type ? `?type=${type}` : "";
    return request<{ usage: UsageOverview | number; type?: string }>(
      `/api/v1/org/billing/usage${params}`
    );
  },

  // Check quota
  checkQuota: (resource: string, amount?: number) => {
    const params = new URLSearchParams({ resource });
    if (amount) params.append("amount", String(amount));
    return request<{ available: boolean }>(`/api/v1/org/billing/quota/check?${params.toString()}`);
  },
};
