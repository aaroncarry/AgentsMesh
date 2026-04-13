import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField } from "@/components/ui/form-field";
import type { TranslationFn } from "./GeneralSettings";

interface AddMemberDirectDialogProps {
  email: string;
  setEmail: (email: string) => void;
  role: "admin" | "member";
  setRole: (role: "admin" | "member") => void;
  adding: boolean;
  onAdd: () => void;
  onClose: () => void;
  t: TranslationFn;
}

export function AddMemberDirectDialog({
  email,
  setEmail,
  role,
  setRole,
  adding,
  onAdd,
  onClose,
  t,
}: AddMemberDirectDialogProps) {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border border-border rounded-lg p-6 w-full max-w-md">
        <h3 className="text-lg font-semibold mb-4">{t("settings.members.addDirectDialog.title")}</h3>
        <p className="text-sm text-muted-foreground mb-4">
          {t("settings.members.addDirectDialog.description")}
        </p>
        <div className="space-y-4">
          <FormField label={t("settings.members.addDirectDialog.emailLabel")} htmlFor="add-email">
            <Input
              id="add-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t("settings.members.addDirectDialog.emailPlaceholder")}
            />
          </FormField>
          <FormField label={t("settings.members.addDirectDialog.roleLabel")} htmlFor="add-role">
            <select
              id="add-role"
              value={role}
              onChange={(e) => setRole(e.target.value as "admin" | "member")}
              className="w-full border border-border rounded px-3 py-2 bg-background"
            >
              <option value="member">{t("settings.members.roleMember")}</option>
              <option value="admin">{t("settings.members.roleAdmin")}</option>
            </select>
          </FormField>
        </div>
        <div className="flex gap-3 mt-6">
          <Button variant="outline" className="flex-1" onClick={onClose}>
            {t("settings.members.addDirectDialog.cancel")}
          </Button>
          <Button
            className="flex-1"
            onClick={onAdd}
            disabled={adding || !email}
          >
            {adding ? t("settings.members.addDirectDialog.adding") : t("settings.members.addDirectDialog.addMember")}
          </Button>
        </div>
      </div>
    </div>
  );
}
