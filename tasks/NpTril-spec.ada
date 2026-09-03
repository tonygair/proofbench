--  <vc-preamble>
package Np_Tril_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   subtype Offset_Type is Integer range -Max_Index .. Max_Index;

   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;

   --  Zero-based row offset of I within A's first dimension.
   function Row_Of (A : Matrix; I : Index_Type) return Integer is
     (I - A'First (1))
   with Pre => I in A'Range (1);

   --  Zero-based column offset of J within A's second dimension.
   function Col_Of (A : Matrix; J : Index_Type) return Integer is
     (J - A'First (2))
   with Pre => J in A'Range (2);
--  </vc-preamble>

--  <vc-spec>
   procedure Tril (A : Matrix; K : Offset_Type; Result : out Matrix) with
     Pre  => A'Length (1) > 0 and then A'Length (2) > 0
             and then Result'First (1) = A'First (1)
             and then Result'Last (1) = A'Last (1)
             and then Result'First (2) = A'First (2)
             and then Result'Last (2) = A'Last (2)
             and then K > -(A'Length (1) - 1)
             and then K < A'Length (2) - 1,
     Post => (for all I in A'Range (1) =>
                (for all J in A'Range (2) =>
                   (if Col_Of (A, J) - Row_Of (A, I) > K
                    then Result (I, J) = 0
                    else Result (I, J) = A (I, J))));

end Np_Tril_Spec;

package body Np_Tril_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Tril (A : Matrix; K : Offset_Type; Result : out Matrix) is
   begin
      pragma Assume (False);
   end Tril;
--  </vc-code>

--  <vc-postamble>
end Np_Tril_Spec;
--  </vc-postamble>
