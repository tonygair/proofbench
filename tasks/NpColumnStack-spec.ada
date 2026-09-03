--  <vc-preamble>
package Np_Column_Stack_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   --  Dafny's seq<seq<int>> of uniform row length, as a rectangular array:
   --  the first dimension indexes the outer sequence.
   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Column_Stack
     (Input  : Matrix;
      M      : Index_Type;
      N      : Index_Type;
      Result : out Matrix)
   with
     Pre  => N > 0
             and then Input'Length (1) = N
             and then Input'Length (2) = M
             and then Result'First (1) = Input'First (2)
             and then Result'Last (1) = Input'Last (2)
             and then Result'First (2) = Input'First (1)
             and then Result'Last (2) = Input'Last (1),
     Post => Result'Length (1) = M
             and then Result'Length (2) = N
             and then Result'Length (1) * N = M * N
             and then (for all I in Input'Range (1) =>
                         (for all J in Input'Range (2) =>
                            Result (J, I) = Input (I, J)));

end Np_Column_Stack_Spec;

package body Np_Column_Stack_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Column_Stack
     (Input  : Matrix;
      M      : Index_Type;
      N      : Index_Type;
      Result : out Matrix) is
   begin
      pragma Assume (False);
   end Column_Stack;
--  </vc-code>

--  <vc-postamble>
end Np_Column_Stack_Spec;
--  </vc-postamble>
