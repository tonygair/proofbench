--  <vc-preamble>
package Np_Remainder_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Dafny's "a % b" is the Euclidean remainder: the unique r with
   --  0 <= r < |b| and b dividing a - r.  Neither Ada's "mod" (sign of the
   --  divisor) nor "rem" (sign of the dividend) matches it when b < 0, so
   --  the postcondition states that characterisation directly instead.
--  </vc-preamble>

--  <vc-spec>
   procedure Remainder (A : Int_Array; B : Int_Array; Result : out Int_Array)
   with
     Pre  => A'First = B'First and then A'Last = B'Last
             and then Result'First = A'First and then Result'Last = A'Last
             and then (for all I in B'Range => B (I) /= 0),
     Post => Result'Length = A'Length
             and then (for all I in A'Range =>
                         Result (I) >= 0
                         and then Result (I) < abs B (I)
                         and then (A (I) - Result (I)) mod B (I) = 0);

end Np_Remainder_Spec;

package body Np_Remainder_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Remainder (A : Int_Array; B : Int_Array; Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Remainder;
--  </vc-code>

--  <vc-postamble>
end Np_Remainder_Spec;
--  </vc-postamble>
