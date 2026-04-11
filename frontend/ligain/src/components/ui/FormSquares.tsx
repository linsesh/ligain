import { View } from 'react-native';
import { FormResult } from '../../utils/standings';
import { colors } from '../../constants/colors';

const FORM_COLORS: Record<FormResult, string> = {
  W: colors.formWin,
  D: colors.formDraw,
  L: colors.formLoss,
};

interface FormSquaresProps {
  form: FormResult[];
  total?: number;
  size?: number;
}

export function FormSquares({ form, total = 5, size = 10 }: FormSquaresProps) {
  const slots = Array.from({ length: total }, (_, i) => form[i] ?? null);
  return (
    <View style={{ flexDirection: 'row', gap: 4 }}>
      {slots.map((result, i) => (
        <View
          key={i}
          style={{
            width: size,
            height: size,
            borderRadius: 2,
            backgroundColor: result ? FORM_COLORS[result] : colors.disabled,
          }}
        />
      ))}
    </View>
  );
}
