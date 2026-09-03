--  <vc-preamble>
package Np_Diagonal_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   subtype Offset_Type is Integer range -Max_Index .. Max_Index;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Diagonal (A : Matrix; K : Offset_Type; Result : out Int_Array)
   with
     Pre  => A'Length (1) > 0
             and then A'Length (2) = A'Length (1)
             and then K > -A'Length (1) and then K < A'Length (1)
             and then (if K > 0
                       then Result'Length = A'Length (1) - K
                       else Result'Length = A'Length (1) + (-K)),
     Post => (if K > 0
              then Result'Length = A'Length (1) - K
                   and then
                     (for all I in 0 .. Result'Length - 1 =>
                        I < A'Length (1)
                        and then I + K < A'Length (1)
                        and then Result (Result'First + I) =
                                 A (A'First (1) + I, A'First (2) + I + K))
              else Result'Length = A'Length (1) + (-K)
                   and then
                     (for all I in 0 .. Result'Length - 1 =>
                        I + (-K) < A'Length (1)
                        and then I < A'Length (1)
                        and then Result (Result'First + I) =
                                 A (A'First (1) + I + (-K), A'First (2) + I)));

end Np_Diagonal_Spec;

package body Np_Diagonal_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Diagonal (A : Matrix; K : Offset_Type; Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Diagonal;
--  </vc-code>

--  <vc-postamble>
end Np_Diagonal_Spec;
--  </vc-postamble>
