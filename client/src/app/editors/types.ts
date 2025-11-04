export type OnChange = (value?: string) => void;

export interface EditorProps {
  initialValue: string | null;
  onChange: OnChange;
}
