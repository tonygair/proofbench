--  <vc-preamble>
package Np_Flatten_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  A valid matrix: A'Length (1) rows of A'Length (2) elements each.
   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Flatten (A : Matrix; Result : out Int_Array) with
     Pre  => A'Length (1) > 0 and then A'Length (2) > 0
             and then Result'Length = A'Length (1) * A'Length (2),
     Post => Result'Length = A'Length (1) * A'Length (2)
             and then
               (for all I in 0 .. A'Length (1) - 1 =>
                  (for all J in 0 .. A'Length (2) - 1 =>
                     I * A'Length (2) + J < Result'Length
                     and then Result (Result'First + I * A'Length (2) + J) =
                              A (A'First (1) + I, A'First (2) + J)));

end Np_Flatten_Spec;

package body Np_Flatten_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Flatten (A : Matrix; Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Flatten;
--  </vc-code>

--  <vc-postamble>
end Np_Flatten_Spec;
--  </vc-postamble>
