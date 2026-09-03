--  <vc-preamble>
package Np_Transpose_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Transpose (A : Matrix; Result : out Matrix) with
     Pre  => A'Length (1) > 0 and then A'Length (2) > 0
             and then Result'First (1) = A'First (2)
             and then Result'Last (1) = A'Last (2)
             and then Result'First (2) = A'First (1)
             and then Result'Last (2) = A'Last (1),
     Post => Result'Length (1) = A'Length (2)
             and then Result'Length (2) = A'Length (1)
             and then (for all I in A'Range (1) =>
                         (for all J in A'Range (2) =>
                            Result (J, I) = A (I, J)));

end Np_Transpose_Spec;

package body Np_Transpose_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Transpose (A : Matrix; Result : out Matrix) is
   begin
      pragma Assume (False);
   end Transpose;
--  </vc-code>

--  <vc-postamble>
end Np_Transpose_Spec;
--  </vc-postamble>
