"use client";

import { useState, useCallback, useEffect, lazy, Suspense } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField, FormRow } from "@/components/ui/form-field";
import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
  ResponsiveDialogBody,
  ResponsiveDialogFooter,
} from "@/components/ui/responsive-dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TicketType, TicketPriority } from "@/lib/api/ticket";
import { ticketApi } from "@/lib/api";
import { organizationApi, OrganizationMember } from "@/lib/api/organization";
import { useAuthStore } from "@/stores/auth";
import { RepositorySelect } from "@/components/common/RepositorySelect";
import { useBreakpoint } from "@/components/layout/useBreakpoint";
import { cn } from "@/lib/utils";
import { ChevronDown, Calendar, Users, X } from "lucide-react";

const BlockEditor = lazy(() => import("@/components/ui/block-editor"));

export interface TicketCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: (ticketId: number, slug: string) => void;
  defaultRepositoryId?: number;
  parentTicketSlug?: string;
}

const typeOptions: { value: TicketType }[] = [
  { value: "task" },
  { value: "bug" },
  { value: "feature" },
  { value: "improvement" },
  { value: "epic" },
];

const priorityOptions: { value: TicketPriority }[] = [
  { value: "urgent" },
  { value: "high" },
  { value: "medium" },
  { value: "low" },
  { value: "none" },
];

interface FormData {
  title: string;
  content: string;
  type: TicketType;
  priority: TicketPriority;
  repositoryId: number | null;
  dueDate: string;
  assigneeIds: number[];
  labels: string[];
}

export function TicketCreateDialog({
  open,
  onOpenChange,
  onCreated,
  defaultRepositoryId,
  parentTicketSlug,
}: TicketCreateDialogProps) {
  const t = useTranslations();
  const { isMobile } = useBreakpoint();
  const { currentOrg } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [moreOptionsOpen, setMoreOptionsOpen] = useState(false);
  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [labelInput, setLabelInput] = useState("");
  const [form, setForm] = useState<FormData>({
    title: "",
    content: "",
    type: "task",
    priority: "medium",
    repositoryId: defaultRepositoryId || null,
    dueDate: "",
    assigneeIds: [],
    labels: [],
  });

  useEffect(() => {
    if (open && currentOrg?.slug && members.length === 0) {
      organizationApi.listMembers(currentOrg.slug)
        .then((res) => setMembers(res.members || []))
        .catch(() => {});
    }
  }, [open, currentOrg?.slug, members.length]);

  const resetForm = useCallback(() => {
    setForm({
      title: "",
      content: "",
      type: "task",
      priority: "medium",
      repositoryId: defaultRepositoryId || null,
      dueDate: "",
      assigneeIds: [],
      labels: [],
    });
    setError(null);
    setMoreOptionsOpen(false);
    setLabelInput("");
  }, [defaultRepositoryId]);

  const handleClose = useCallback(() => {
    onOpenChange(false);
    resetForm();
  }, [onOpenChange, resetForm]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!form.title.trim()) {
      setError(t("tickets.createDialog.titleRequired"));
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await ticketApi.create({
        repositoryId: form.repositoryId || undefined,
        title: form.title.trim(),
        content: form.content || undefined,
        type: form.type,
        priority: form.priority,
        parentSlug: parentTicketSlug,
        assigneeIds: form.assigneeIds.length > 0 ? form.assigneeIds : undefined,
        labels: form.labels.length > 0 ? form.labels : undefined,
      });

      onCreated?.(response.id, response.slug);
      handleClose();
    } catch (err: unknown) {
      console.error("Failed to create ticket:", err);
      setError(err instanceof Error ? err.message : t("tickets.createDialog.createFailed"));
    } finally {
      setLoading(false);
    }
  };

  const updateField = <K extends keyof FormData>(key: K, value: FormData[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (error) setError(null);
  };

  const toggleAssignee = (userId: number) => {
    setForm((prev) => ({
      ...prev,
      assigneeIds: prev.assigneeIds.includes(userId)
        ? prev.assigneeIds.filter((id) => id !== userId)
        : [...prev.assigneeIds, userId],
    }));
  };

  const addLabel = (label: string) => {
    const trimmed = label.trim();
    if (trimmed && !form.labels.includes(trimmed)) {
      updateField("labels", [...form.labels, trimmed]);
    }
    setLabelInput("");
  };

  const removeLabel = (label: string) => {
    updateField("labels", form.labels.filter((l) => l !== label));
  };

  const handleLabelKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addLabel(labelInput);
    } else if (e.key === "Backspace" && !labelInput && form.labels.length > 0) {
      removeLabel(form.labels[form.labels.length - 1]);
    }
  };

  const dialogTitle = parentTicketSlug
    ? t("tickets.createDialog.createSubTicket")
    : t("tickets.createDialog.title");

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent className="max-w-lg" title={dialogTitle}>
        <ResponsiveDialogHeader>
          <ResponsiveDialogTitle>{dialogTitle}</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0">
          <ResponsiveDialogBody className="space-y-4">
            {/* Title */}
            <FormField
              label={t("tickets.createDialog.titleLabel")}
              htmlFor="ticket-title"
              required
            >
              <Input
                id="ticket-title"
                placeholder={t("tickets.createDialog.titlePlaceholder")}
                value={form.title}
                onChange={(e) => updateField("title", e.target.value)}
                autoFocus
              />
            </FormField>

            {/* Type & Priority */}
            <FormRow>
              <FormField label={t("tickets.filters.type")} htmlFor="ticket-type">
                <Select
                  value={form.type}
                  onValueChange={(val) => updateField("type", val as TicketType)}
                >
                  <SelectTrigger id="ticket-type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {typeOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {t(`tickets.type.${opt.value}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>

              <FormField label={t("tickets.filters.priority")} htmlFor="ticket-priority">
                <Select
                  value={form.priority}
                  onValueChange={(val) => updateField("priority", val as TicketPriority)}
                >
                  <SelectTrigger id="ticket-priority">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {priorityOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {t(`tickets.priority.${opt.value}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            </FormRow>

            {/* Repository */}
            <FormField label={t("tickets.createDialog.repository")} htmlFor="ticket-repo">
              <RepositorySelect
                value={form.repositoryId}
                onChange={(value) => updateField("repositoryId", value)}
                placeholder={t("tickets.createDialog.selectRepository")}
              />
            </FormField>

            {/* Content */}
            <FormField label={t("tickets.createDialog.content")}>
              <div className={cn(
                "border border-input rounded-md overflow-hidden bg-card",
                isMobile ? "min-h-[100px]" : "min-h-[150px]"
              )}>
                <Suspense fallback={<div className={cn("animate-pulse bg-muted", isMobile ? "h-[100px]" : "h-[150px]")} />}>
                  <BlockEditor
                    initialContent={form.content}
                    onChange={(content) => updateField("content", content)}
                    editable={true}
                  />
                </Suspense>
              </div>
            </FormField>

            {/* More options toggle */}
            <button
              type="button"
              className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
              onClick={() => setMoreOptionsOpen(!moreOptionsOpen)}
            >
              <ChevronDown className={cn("h-4 w-4 transition-transform", moreOptionsOpen && "rotate-180")} />
              {t("tickets.createDialog.moreOptions") || "More options"}
            </button>

            {moreOptionsOpen && (
              <div className="space-y-4 pt-1 border-t border-border">
                {/* Due date */}
                <FormField label={t("tickets.detail.dueDate")} htmlFor="ticket-due-date">
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                    <Input
                      id="ticket-due-date"
                      type="date"
                      value={form.dueDate}
                      onChange={(e) => updateField("dueDate", e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </FormField>

                {/* Assignees */}
                <FormField label={t("tickets.detail.assignees")}>
                  <div className="space-y-2">
                    {members.length > 0 ? (
                      <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto">
                        {members.map((member) => {
                          const isSelected = form.assigneeIds.includes(member.user_id);
                          return (
                            <button
                              key={member.user_id}
                              type="button"
                              className={cn(
                                "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs border transition-colors",
                                isSelected
                                  ? "border-primary bg-primary/10 text-primary"
                                  : "border-border hover:border-primary/50 text-muted-foreground hover:text-foreground"
                              )}
                              onClick={() => toggleAssignee(member.user_id)}
                            >
                              {member.user?.avatar_url ? (
                                /* eslint-disable-next-line @next/next/no-img-element */
                                <img
                                  src={member.user.avatar_url}
                                  alt=""
                                  className="w-4 h-4 rounded-full"
                                />
                              ) : (
                                <Users className="w-3 h-3" />
                              )}
                              {member.user?.name || member.user?.username || member.user?.email}
                            </button>
                          );
                        })}
                      </div>
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        {t("tickets.detail.noAssignees")}
                      </p>
                    )}
                  </div>
                </FormField>

                {/* Labels */}
                <FormField label={t("tickets.detail.labels")}>
                  <div className="flex flex-wrap items-center gap-1.5 border border-input rounded-md px-2 py-1.5 min-h-[38px] focus-within:ring-2 focus-within:ring-primary/20">
                    {form.labels.map((label) => (
                      <span
                        key={label}
                        className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-primary/10 text-primary text-xs"
                      >
                        {label}
                        <button
                          type="button"
                          onClick={() => removeLabel(label)}
                          className="hover:text-destructive"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </span>
                    ))}
                    <input
                      type="text"
                      value={labelInput}
                      onChange={(e) => setLabelInput(e.target.value)}
                      onKeyDown={handleLabelKeyDown}
                      onBlur={() => { if (labelInput.trim()) addLabel(labelInput); }}
                      placeholder={form.labels.length === 0 ? (t("tickets.createDialog.labelsPlaceholder") || "Add labels...") : ""}
                      className="flex-1 min-w-[80px] bg-transparent text-sm outline-none placeholder:text-muted-foreground"
                    />
                  </div>
                </FormField>
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded-md">
                {error}
              </div>
            )}
          </ResponsiveDialogBody>

          <ResponsiveDialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={loading}
              className="w-full sm:w-auto"
            >
              {t("common.cancel")}
            </Button>
            <Button type="submit" loading={loading} className="w-full sm:w-auto">
              {t("tickets.createDialog.submit")}
            </Button>
          </ResponsiveDialogFooter>
        </form>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}

export default TicketCreateDialog;
